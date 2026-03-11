package collector

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
)

type BatteryStats struct {
	Present     bool
	Capacity    int     // Percentage
	Status      string  // Charging, Discharging, etc.
	Power       float64 // Watts (discharging/charging rate)
	EnergyNow   float64 // Wh
	EnergyFull  float64 // Wh
	Voltage     float64 // V
}

func GetBatteryStats() BatteryStats {
	var stats BatteryStats
	
	powerSupplies, _ := os.ReadDir("/sys/class/power_supply")
	var path string
	found := false
	for _, ps := range powerSupplies {
		if strings.HasPrefix(ps.Name(), "BAT") {
			path = "/sys/class/power_supply/" + ps.Name()
			found = true
			break
		}
	}
	
	if !found {
		return stats
	}

	stats.Present = true
	
	capacity, _ := os.ReadFile(path + "/capacity")
	stats.Capacity, _ = strconv.Atoi(strings.TrimSpace(string(capacity)))

	status, _ := os.ReadFile(path + "/status")
	stats.Status = strings.TrimSpace(string(status))

	// Power in microwatts
	powerNow, _ := os.ReadFile(path + "/power_now")
	if len(powerNow) > 0 {
		pMicros, _ := strconv.ParseFloat(strings.TrimSpace(string(powerNow)), 64)
		stats.Power = pMicros / 1000000.0
	} else {
		// Fallback for some systems using current_now * voltage_now
		currNow, _ := os.ReadFile(path + "/current_now")
		voltNow, _ := os.ReadFile(path + "/voltage_now")
		if len(currNow) > 0 && len(voltNow) > 0 {
			cMicros, _ := strconv.ParseFloat(strings.TrimSpace(string(currNow)), 64)
			vMicros, _ := strconv.ParseFloat(strings.TrimSpace(string(voltNow)), 64)
			stats.Power = (cMicros * vMicros) / 1000000000000.0 // (uA * uV) -> Watts
		}
	}

	energyNow, _ := os.ReadFile(path + "/energy_now")
	if len(energyNow) > 0 {
		eMicros, _ := strconv.ParseFloat(strings.TrimSpace(string(energyNow)), 64)
		stats.EnergyNow = eMicros / 1000000.0
	}

	energyFull, _ := os.ReadFile(path + "/energy_full")
	if len(energyFull) > 0 {
		eMicros, _ := strconv.ParseFloat(strings.TrimSpace(string(energyFull)), 64)
		stats.EnergyFull = eMicros / 1000000.0
	}

	voltageNow, _ := os.ReadFile(path + "/voltage_now")
	if len(voltageNow) > 0 {
		vMicros, _ := strconv.ParseFloat(strings.TrimSpace(string(voltageNow)), 64)
		stats.Voltage = vMicros / 1000000.0
	}

	return stats
}

func GetBatteryState(ctx context.Context, ticker <-chan time.Time) chan BatteryStats {
	output := make(chan BatteryStats)
	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				output <- GetBatteryStats()
			}
		}
	}()
	return output
}
