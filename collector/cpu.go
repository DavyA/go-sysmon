package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CpuCore rappresenta lo stato di un singolo core
type CpuCore struct {
	ID    string
	Usage float64 // Percentuale 0-100
	Temp  float64 // Temperatura in Celsius
}

type CpuStats struct {
	CpuCores     []CpuCore
	PackagePower float64 // Watts
}

// Struct per mantenere lo stato precedente (necessario per calcolare il delta)
type CpuTracker struct {
	mu           sync.Mutex
	prevIdle     map[string]float64
	prevTotal    map[string]float64
	prevEnergyUJ int64
	prevTime     time.Time
}

func NewCpuTracker() *CpuTracker {
	return &CpuTracker{
		prevIdle:  make(map[string]float64),
		prevTotal: make(map[string]float64),
		prevTime:  time.Now(),
	}
}

// GetCpuModel legge il nome della CPU (es. AMD Ryzen 7 3700X)
func GetCpuModel() string {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "Unknown CPU"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return "Unknown CPU"
}

// GetCPUTemp legge la temperatura dal file di sistema (funziona su AMD/Intel)
func (t *CpuTracker) GetCPUTemp(ctx context.Context, ticker <-chan time.Time) chan map[string]float64 {
	output := make(chan map[string]float64)
	temps := make(map[string]float64)
	hwmons, _ := filepath.Glob("/sys/class/hwmon/hwmon*")

	go func() {
		defer close(output)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				for _, path := range hwmons {
					// Verifichiamo se è il sensore della CPU leggendo il file 'name'
					name, _ := os.ReadFile(filepath.Join(path, "name"))
					if strings.TrimSpace(string(name)) != "k10temp" && strings.TrimSpace(string(name)) != "coretemp" {
						continue
					}

					// Cerchiamo tutti i file input in questa cartella
					files, _ := filepath.Glob(filepath.Join(path, "temp*_input"))
					for _, file := range files {
						// Leggiamo la label associata (es. temp1_input -> temp1_label)
						labelFile := strings.Replace(file, "_input", "_label", 1)
						label, _ := os.ReadFile(labelFile)

						// Leggiamo il valore
						valRaw, _ := os.ReadFile(file)
						valInt, _ := strconv.Atoi(strings.TrimSpace(string(valRaw)))

						labelText := strings.TrimSpace(string(label))
						if labelText == "" {
							labelText = filepath.Base(file)
						}

						temps[labelText] = float64(valInt) / 1000.0
					}
				}
				output <- temps
			}
		}
	}()
	return output
}

// GetPerCoreUsage calcola l'uso CPU per ogni core
func (t *CpuTracker) GetCPUUsage(ctx context.Context, ticker <-chan time.Time) chan CpuStats {
	output := make(chan CpuStats)
	go func() {
		defer close(output)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				file, err := os.ReadFile("/proc/stat")
				if err != nil {
					return
				}

				lines := strings.Split(string(file), "\n")
				var cpuStats CpuStats

				for _, line := range lines {
					// Cerchiamo righe che iniziano con "cpu0", "cpu1", ecc. (ma non "cpu " totale)
					if strings.HasPrefix(line, "cpu") && !strings.HasPrefix(line, "cpu ") {
						fields := strings.Fields(line)
						cpuID := fields[0]

						// Parsing dei valori (User, Nice, System, Idle, Iowait, Irq, Softirq, Steal)
						var total float64
						var idle float64

						for i, valStr := range fields[1:] {
							val, _ := strconv.ParseFloat(valStr, 64)
							total += val
							if i == 3 { // Il 4° campo è IDLE
								idle = val
							}
						}

						// Calcolo del Delta rispetto alla lettura precedente
						t.mu.Lock()
						prevIdle := t.prevIdle[cpuID]
						prevTotal := t.prevTotal[cpuID]
						deltaTotal := total - prevTotal
						deltaIdle := idle - prevIdle

						usage := 0.0
						if deltaTotal > 0 {
							usage = ((deltaTotal - deltaIdle) / deltaTotal) * 100
						}

						t.prevIdle[cpuID] = idle
						t.prevTotal[cpuID] = total
						t.mu.Unlock()

						cpuStats.CpuCores = append(cpuStats.CpuCores, CpuCore{ID: cpuID, Usage: usage})
					}
				}

				// Calcoliamo Watt reali
				t.mu.Lock()
				now := time.Now()

				// 1. Try Intel/AMD RAPL (requires root or permissions)
				raplData, err := os.ReadFile("/sys/class/powercap/intel-rapl:0/energy_uj")
				if err == nil {
					currUJ, _ := strconv.ParseInt(strings.TrimSpace(string(raplData)), 10, 64)
					if t.prevEnergyUJ > 0 && currUJ > t.prevEnergyUJ {
						diffUJ := currUJ - t.prevEnergyUJ
						diffSec := now.Sub(t.prevTime).Seconds()
						if diffSec > 0 {
							cpuStats.PackagePower = (float64(diffUJ) / 1000000.0) / diffSec
						}
					}
					t.prevEnergyUJ = currUJ
				} else {
					// 2. Fallback to hwmon (amdgpu APU PPT or coretemp)
					hwmons, _ := filepath.Glob("/sys/class/hwmon/hwmon*")
					for _, path := range hwmons {
						name, _ := os.ReadFile(filepath.Join(path, "name"))
						n := strings.TrimSpace(string(name))
						if n == "k10temp" || n == "coretemp" || n == "amdgpu" {
							pBytes, err := os.ReadFile(filepath.Join(path, "power1_average"))
							if err != nil {
								pBytes, _ = os.ReadFile(filepath.Join(path, "power1_input"))
							}
							if len(pBytes) > 0 {
								var microwatts float64
								fmt.Sscanf(strings.TrimSpace(string(pBytes)), "%f", &microwatts)
								if microwatts > 0 {
									cpuStats.PackagePower = microwatts / 1000000.0
									break // Trovato!
								}
							}
						}
					}
				}

				t.prevTime = now
				t.mu.Unlock()

				output <- cpuStats
			}
		}
	}()
	return output
}
