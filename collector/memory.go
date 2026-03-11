package collector

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
)

type MemState struct {
	MemTotal   uint64
	MemUsed    uint64
	MemFree    uint64
	SwapTotal  uint64
	SwapFree   uint64
	SwapCached uint64
	SwapUsed   uint64
}

// GetSystemState reads the memory information from /proc/meminfo
func GetSystemState(ctx context.Context, ticker <-chan time.Time) chan MemState {
	output := make(chan MemState)
	var memState MemState
	var lines []string
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				data, err := os.ReadFile("/proc/meminfo")
				if err != nil {
					output <- MemState{}
				}
				lines = strings.Split(string(data), "\n")

				for _, line := range lines {
					fields := strings.Fields(line)
					if len(fields) < 2 {
						continue
					}
					val, _ := strconv.Atoi(fields[1])

					switch fields[0] {
					case "MemTotal:":
						memState.MemTotal = uint64(val / 1024)
					case "MemAvailable:":
						memState.MemFree = uint64(val / 1024)
					case "SwapTotal:":
						memState.SwapTotal = uint64(val / 1024)
					case "SwapFree:":
						memState.SwapFree = uint64(val / 1024)
					case "SwapCached:":
						memState.SwapCached = uint64(val / 1024)
					}
				}
				memState.SwapUsed = memState.SwapTotal - memState.SwapFree - memState.SwapCached
				memState.MemUsed = memState.MemTotal - memState.MemFree
				output <- memState
			}
		}
	}()
	return output
}
