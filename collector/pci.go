package collector

import (
	"fmt"
	"os/exec"
	"strings"
)

type PciDevice struct {
	Address string
	Name    string
}

// GetGraphicsDevices usa lspci per trovare GPU e NPU
func GetGraphicsDevices() ([]PciDevice, error) {
	cmd := exec.Command("lspci", "-mm")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var devices []PciDevice
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Split on quotes handling
		var parts []string
		curr := ""
		inQuotes := false
		for i := 0; i < len(line); i++ {
			c := line[i]
			if c == '"' {
				inQuotes = !inQuotes
			} else if c == ' ' && !inQuotes {
				parts = append(parts, curr)
				curr = ""
			} else {
				curr += string(c)
			}
		}
		if curr != "" {
			parts = append(parts, curr)
		}

		if len(parts) < 4 {
			continue
		}

		slot := parts[0]
		class := parts[1]
		vendor := parts[2]
		device := parts[3]

		lowerClass := strings.ToLower(class)
		if strings.Contains(lowerClass, "vga") ||
			strings.Contains(lowerClass, "3d") ||
			strings.Contains(lowerClass, "processing") ||
			strings.Contains(lowerClass, "display") {

			fullDesc := fmt.Sprintf("[%s] %s %s", class, vendor, device)
			devices = append(devices, PciDevice{Address: slot, Name: fullDesc})
		}
	}
	return devices, nil
}
