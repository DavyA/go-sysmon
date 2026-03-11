package collector

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Process struct {
	PID     int
	User    string
	Name    string
	CPU     float64
	MemRSS  uint64 // MiB
	Status  string
	CmdLine string
	IsGPU   bool
}

type procSample struct {
	ticks uint64
	time  time.Time
}

type ProcessTracker struct {
	mu      sync.Mutex
	samples map[int]procSample
}

func NewProcessTracker() *ProcessTracker {
	return &ProcessTracker{
		samples: make(map[int]procSample),
	}
}

func (pt *ProcessTracker) GetAllProcesses() []Process {
	pids, _ := filepath.Glob("/proc/[0-9]*")
	gpuPids := GetGpuPIDs() // Get GPU map Once

	var processes []Process
	now := time.Now()

	pt.mu.Lock()
	newSamples := make(map[int]procSample)

	for _, p := range pids {
		pidStr := filepath.Base(p)
		pid, err := strconv.Atoi(pidStr)
		if err != nil { continue }

		// NVIDIA Fallback: Check /dev/nvidia in fd (only for very active or suspicious ones)
		isGPU := gpuPids[pid]
		if !isGPU {
			// Fast check: check if it has MANY fds? No, just stick to sysfs for now.
		}

		// 1. CPU (utime + stime from /proc/[pid]/stat)
		// stat fields are 1-based. utime is 14, stime 15.
		data, err := os.ReadFile(filepath.Join(p, "stat"))
		if err != nil { continue }

		fields := strings.Fields(string(data))
		if len(fields) < 15 { continue }

		utime, _ := strconv.ParseUint(fields[13], 10, 64)
		stime, _ := strconv.ParseUint(fields[14], 10, 64)
		totalTicks := utime + stime

		cpuPerc := 0.0
		if prev, exists := pt.samples[pid]; exists {
			dt := now.Sub(prev.time).Seconds()
			if dt > 0 {
				// Assume 100 ticks per second (Standard Linux CLK_TCK)
				cpuPerc = (float64(totalTicks-prev.ticks) / 100.0) / dt * 100.0
			}
		}
		newSamples[pid] = procSample{ticks: totalTicks, time: now}

		// 2. Memory (VmRSS from /proc/[pid]/status)
		// 3. User
		statusData, err := os.ReadFile(filepath.Join(p, "status"))
		if err != nil { continue }

		var rss uint64
		var uid string
		var state string
		name := fields[1] // (name) with parens
		name = strings.Trim(name, "()")

		lines := strings.Split(string(statusData), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "VmRSS:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					rss, _ = strconv.ParseUint(f[1], 10, 64)
				}
			}
			if strings.HasPrefix(line, "Uid:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					uid = f[1]
				}
			}
			if strings.HasPrefix(line, "State:") {
				f := strings.Fields(line)
				if len(f) >= 2 {
					state = f[1]
				}
			}
		}

		userName := uid
		if u, err := user.LookupId(uid); err == nil {
			userName = u.Username
		}

		// 4. Command Line
		cmdLineData, _ := os.ReadFile(filepath.Join(p, "cmdline"))
		cmdLine := string(cmdLineData)
		cmdLine = strings.ReplaceAll(cmdLine, "\x00", " ")
		if cmdLine == "" { cmdLine = name }

		processes = append(processes, Process{
			PID:     pid,
			User:    userName,
			Name:    name,
			CPU:     cpuPerc,
			MemRSS:  rss / 1024,
			Status:  state,
			CmdLine: cmdLine,
			IsGPU:   isGPU,
		})
	}

	pt.samples = newSamples
	pt.mu.Unlock()

	// Default sort by CPU
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPU > processes[j].CPU
	})

	return processes
}

func (pt *ProcessTracker) GetProcessState(ctx context.Context, ticker <-chan time.Time) chan []Process {
	output := make(chan []Process)
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done(): return
			case <-ticker:
				output <- pt.GetAllProcesses()
			}
		}
	}()
	return output
}

func KillProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil { return err }
	return proc.Kill()
}
