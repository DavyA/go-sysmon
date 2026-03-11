package collector

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
)

type NetInterface struct {
	Name    string
	RxBytes uint64
	TxBytes uint64
	RxSpeed float64
	TxSpeed float64
}

func GetNetTraffic() ([]NetInterface, error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var ifaces []NetInterface

	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.Split(line, ":")
		name := strings.TrimSpace(parts[0])

		if name == "lo" {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}

		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)

		if rx > 0 || tx > 0 {
			ifaces = append(ifaces, NetInterface{Name: name, RxBytes: rx, TxBytes: tx})
		}
	}
	return ifaces, nil
}

func GetNetSpeed(ctx context.Context, ticker <-chan time.Time) chan []NetInterface {
	output := make(chan []NetInterface)
	prevIfaces, _ := GetNetTraffic()
	go func() {
		defer close(output)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				ifaces, _ := GetNetTraffic()

				for i := range ifaces {
					for _, prevIface := range prevIfaces {
						if ifaces[i].Name == prevIface.Name {
							// Calcolo delta
							ifaces[i].RxSpeed = float64(ifaces[i].RxBytes-prevIface.RxBytes) * 8 / 1024 / 1024
							ifaces[i].TxSpeed = float64(ifaces[i].TxBytes-prevIface.TxBytes) * 8 / 1024 / 1024
						}
					}
				}

				// INVIO E AGGIORNAMENTO (Fondamentale!)
				output <- ifaces
				prevIfaces = ifaces
			}
		}
	}()

	return output
}
