package collector

import (
	"bufio"
	"context"
	"os"
	"strings"
	"syscall"
	"time"
)

type Partition struct {
	MountPoint string
	Used      uint64 // MiB
	Total     uint64 // MiB
	Percent   float64
}

func GetDiskUsage() []Partition {
	var parts []Partition
	file, err := os.Open("/proc/mounts")
	if err != nil { return nil }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 { continue }
		
		device := fields[0]
		mount := fields[1]
		fsType := fields[2]

		// Skip common non-real filesystems
		if !strings.HasPrefix(device, "/dev/") || fsType == "squashfs" || fsType == "tmpfs" {
			continue
		}

		var stat syscall.Statfs_t
		if err := syscall.Statfs(mount, &stat); err == nil {
			total := stat.Blocks * uint64(stat.Bsize)
			free := stat.Bfree * uint64(stat.Bsize)
			used := total - free

			parts = append(parts, Partition{
				MountPoint: mount,
				Used:      used / 1024 / 1024,
				Total:     total / 1024 / 1024,
				Percent:   float64(used) / float64(total) * 100,
			})
		}
	}
	return parts
}

func GetDiskState(ctx context.Context, ticker <-chan time.Time) chan []Partition {
	output := make(chan []Partition)
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done(): return
			case <-ticker:
				output <- GetDiskUsage()
			}
		}
	}()
	return output
}
