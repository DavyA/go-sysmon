package utils

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sysmon/collector"
	"time"
)

func getOSName() string {
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
			}
		}
	}
	return "Linux"
}

func getKernelRelease() string {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err == nil {
		return strings.TrimSpace(string(data))
	}
	return "Unknown"
}

func getUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			var uptimeSec float64
			fmt.Sscanf(fields[0], "%f", &uptimeSec)
			d := time.Duration(uptimeSec) * time.Second
			hours := int(d.Hours())
			minutes := int(d.Minutes()) % 60
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
	}
	return "Unknown"
}

// SystemState mantiene gli ultimi dati raccolti in modo thread-safe
type SystemState struct {
	Mu          sync.RWMutex
	Hostname    string
	CPUStats    collector.CpuStats
	CPUTemp     map[string]float64
	NetStats    []collector.NetInterface
	Memory      collector.MemState
	Model       string
	PCI         []collector.PciDevice
	GPUs        []collector.GPUStatus
	CPUHistory  []float64
	GPUHistory  map[string][]float64
	CurrentView ViewMode
	cachedOS    string
	cachedKrn   string
}

type ViewMode int

const (
	ViewDashboard ViewMode = iota
	ViewCPU
	ViewGPU
	ViewPIDs
	ViewInfo
)

func RenderNeofetch(s *SystemState) {
	s.Mu.Lock()
	if s.cachedOS == "" {
		s.cachedOS = getOSName()
		s.cachedKrn = getKernelRelease()
	}
	osName := s.cachedOS
	kernel := s.cachedKrn
	s.Mu.Unlock()

	uptime := getUptime()

	// Leggiamo la RAM
	s.Mu.RLock()
	usedRam := s.Memory.MemUsed
	totalRam := s.Memory.MemTotal
	cpuMod := s.Model
	s.Mu.RUnlock()

	logoObj := GetOSLogo(osName, kernel)
	lines := strings.Split(logoObj.Art, "\n")

	// Rimuoviamo la prima riga vuota se inizia con un \n nudo
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// Aggiungiamo padding vuoto se il logo ha meno di 6 righe
	for len(lines) < 6 {
		lines = append(lines, "")
	}

	userName := os.Getenv("USER")
	if userName == "" {
		userName = "user"
	}
	shellPath := os.Getenv("SHELL")
	shellParts := strings.Split(shellPath, "/")
	shell := shellParts[len(shellParts)-1]
	if shell == "" {
		shell = "bash"
	}

	c := logoObj.Color
	b := Bold
	r := Reset

	// Costruiamo user@hostname colorato (con il colore del logo)
	userHost := fmt.Sprintf("%s%s%s%s@%s%s%s%s", c, b, userName, r, c, b, s.Hostname, r)
	separator := fmt.Sprintf("%s%s%s", c, strings.Repeat("-", len(userName)+len(s.Hostname)+1), r)

	// Righe di info formatte alla Neofetch/Archey
	infos := []string{
		userHost,
		separator,
		fmt.Sprintf("%s%sOS:%s %s", c, b, r, osName),
		fmt.Sprintf("%s%sKernel:%s %s", c, b, r, kernel),
		fmt.Sprintf("%s%sUptime:%s %s", c, b, r, uptime),
		fmt.Sprintf("%s%sShell:%s %s", c, b, r, shell),
		fmt.Sprintf("%s%sCPU:%s %s", c, b, r, strings.TrimSpace(cpuMod)),
		fmt.Sprintf("%s%sMemory:%s %d MiB / %d MiB", c, b, r, usedRam, totalRam),
		"",
		"\033[40m   \033[41m   \033[42m   \033[43m   \033[44m   \033[45m   \033[46m   \033[47m   \033[0m",
	}

	fmt.Printf("\r\n") // Spazio sopra
	for i := 0; i < len(lines); i++ {
		infoStr := ""
		if i < len(infos) {
			infoStr = "  " + infos[i]
		}

		fmt.Printf("%s%-25s%s %s\r\n", logoObj.Color, lines[i], Reset, infoStr)
	}

	// Se infos è più lungo del logo
	for i := len(lines); i < len(infos); i++ {
		fmt.Printf("%-25s   %s\r\n", "", infos[i])
	}
}
