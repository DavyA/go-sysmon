package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	// IMPORTANTE: Questo deve matchare il nome in go.mod
	"sysmon/collector"
	"sysmon/utils"
)

const Millisecond = 1000

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		// Esegui la pulizia manualmente prima di uscire
		fmt.Print("\033[?25h\033[?1049l")
		os.Exit(0)
	}()

	// 1. Entra nell'Alternate Buffer e Nascondi il Cursore
	fmt.Print("\033[?1049h") // Entra in alternate screen
	fmt.Print("\033[?25l")   // Nasconde il cursore (molto più pulito!)

	// 2. Assicurati di uscire pulito quando chiudi il programma
	defer func() {
		fmt.Print("\033[?25h")   // Mostra di nuovo il cursore
		fmt.Print("\033[?1049l") // Torna al buffer principale
	}()

	cpuTicker, cpuTicker_stop := immediateTicker(Millisecond * time.Millisecond)
	tempTicker, tempTicker_stop := immediateTicker(Millisecond * time.Millisecond)
	netTicker, netTicker_stop := immediateTicker(Millisecond * time.Millisecond)
	displayTicker, displayTicker_stop := immediateTicker(Millisecond * time.Millisecond)
	memTicker, memTicker_stop := immediateTicker(Millisecond * time.Millisecond)
	gpuTicker, gpuTicker_stop := immediateTicker(Millisecond * time.Millisecond)

	defer cpuTicker_stop()
	defer tempTicker_stop()
	defer netTicker_stop()
	defer displayTicker_stop()
	defer memTicker_stop()
	defer gpuTicker_stop()

	cpuTracker := collector.NewCpuTracker()
	cpuChan := cpuTracker.GetCPUUsage(ctx, cpuTicker)
	tempChan := cpuTracker.GetCPUTemp(ctx, tempTicker)
	netChan := collector.GetNetSpeed(ctx, netTicker)
	memChan := collector.GetSystemState(ctx, memTicker)
	gpuChan := collector.GetGPUState(ctx, gpuTicker)

	// Inizializziamo lo stato
	state := &utils.SystemState{
		Model: collector.GetCpuModel(),
	}
	state.PCI, _ = collector.GetGraphicsDevices()
	state.Hostname = getHostname()

	// 1. Goroutine per aggiornamento CPU
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case stats := <-cpuChan:
				state.Mu.Lock()
				state.CPUStats = stats
				// Calcola la media totale dei core
				var total float64
				for _, c := range stats.CpuCores {
					total += c.Usage
				}
				avg := total / float64(len(stats.CpuCores))

				// Mantieni la cronologia (es. ultimi 100 campioni)
				state.CPUHistory = append(state.CPUHistory, avg)
				if len(state.CPUHistory) > 100 {
					state.CPUHistory = state.CPUHistory[1:]
				}
				state.Mu.Unlock()
			}
		}
	}()

	// 2. Goroutine per aggiornamento Rete
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case nets := <-netChan:
				state.Mu.Lock()
				state.NetStats = nets
				state.Mu.Unlock()
			}
		}
	}()

	// 3. Goroutine per aggiornamento temperatura
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case temps := <-tempChan:
				state.Mu.Lock()
				state.CPUTemp = temps
				state.Mu.Unlock()
			}
		}
	}()

	// 4. Goroutine per aggiornamento memoria
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case mem := <-memChan:
				state.Mu.Lock()
				state.Memory = mem
				state.Mu.Unlock()
			}
		}
	}()

	// 5. Goroutine per aggiornamento GPU
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case gpus := <-gpuChan:
				state.Mu.Lock()
				state.GPUs = gpus
				if state.GPUHistory == nil {
					state.GPUHistory = make(map[string][]float64)
				}
				for _, g := range gpus {
					state.GPUHistory[g.ID] = append(state.GPUHistory[g.ID], g.Utilization)
					if len(state.GPUHistory[g.ID]) > 100 { // Max 100 steps
						state.GPUHistory[g.ID] = state.GPUHistory[g.ID][1:]
					}
				}
				state.Mu.Unlock()
			}
		}
	}()

	// Nel Main:
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	var RedrawChan chan bool
	RedrawChan = make(chan bool, 1)
	go utils.HandleResize(sigChan, RedrawChan)
	go utils.HandleInput(state, RedrawChan)

	// Nel main
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		// Esegui la pulizia manualmente prima di uscire
		fmt.Print("\033[?25h\033[?1049l")
		os.Exit(0)
	}()

	//Facciamo un primo update dei valori

	fmt.Println("Starting Go-Sysmon Advanced System Monitor...")

	for {
		select {
		case <-RedrawChan:
			clearScreen()
			state.Render()
		case <-displayTicker:
			clearScreen()
			state.Render()

		}
	}
	/*
		fmt.Println("=============================================================")
		fmt.Printf(" HOST: %s | CPU: %s | TIME: %s\n", state.Hostname, state.Model, t.Format("15:04:05.000"))
		fmt.Println("=============================================================")

		if len(state.PCI) > 0 {
			fmt.Println(" [PCI / GPU]")
			for _, gpu := range state.PCI {
				fmt.Printf("  -> %s: %s\n", gpu.Address, gpu.Name)
			}
			fmt.Println("-------------------------------------------------------------")
		}

		// CPU Temp
		state.Mu.RLock()
		cpuTemp := state.CPUTemp
		state.Mu.RUnlock()
		fmt.Println(" [CPU TEMPERATURE]")
		for name, temp := range cpuTemp {
			fmt.Printf("  -> %s: %.2f°C\n", name, temp)
		}
		fmt.Println("-------------------------------------------------------------")

		// CPU Cores
		state.Mu.RLock()
		cpuStats := state.CPUStats
		state.Mu.RUnlock()
		fmt.Println(" [CPU CORES LOAD]")
		for i, core := range cpuStats.CpuCores {
			bar := createBar(int(core.Usage))
			fmt.Printf(" %-4s [%-10s] %3.0f%% ", core.ID, bar, core.Usage)
			if (i+1)%2 == 0 {
				fmt.Println()
			}
		}
		if len(cpuStats.CpuCores)%2 != 0 {
			fmt.Println()
		}
		fmt.Println("-------------------------------------------------------------")

		fmt.Printf(" [ CPU LOAD HISTORY ]\n")
		top, middle, bottom := utils.CreateSolidGraph(state.CPUHistory, 40)
		fmt.Printf(" %s\n", top)
		fmt.Printf(" %s\n", middle)
		fmt.Printf(" %s\n", bottom)
		fmt.Println("-------------------------------------------------------------")

		// RAM
		state.Mu.RLock()
		mem := state.Memory
		state.Mu.RUnlock()
		fmt.Printf(" [MEMORY] Used: %d / %d MB\t%.2f%%\n", mem.MemUsed, mem.MemTotal, float64(mem.MemUsed)/float64(mem.MemTotal)*100)
		fmt.Printf(" [SWAP] Used: %d / %d MB\t%.2f%%\n", mem.SwapUsed, mem.SwapTotal, float64(mem.SwapUsed)/float64(mem.SwapTotal)*100)

		// NET

		state.Mu.RLock()
		nets := state.NetStats
		state.Mu.RUnlock()
		fmt.Println("-------------------------------------------------------------")
		fmt.Println(" [NETWORK TRAFFIC]")
		for _, nic := range nets {
			fmt.Printf("  -> %-10s RX Speed: %.2f MB/s | TX Speed: %.2f MB/s\n",
				nic.Name, nic.RxSpeed, nic.TxSpeed)
		}
		fmt.Println("=============================================================")
	*/
}

func getHostname() string {
	name, _ := os.Hostname()
	return name
}

func createBar(percent int) string {
	width := 10
	filled := (percent * width) / 100
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "#"
		} else {
			bar += "."
		}
	}
	return bar
}

func clearScreen() {
	fmt.Print("\033[H\033[J")
}

func immediateTicker(d time.Duration) (<-chan time.Time, func()) {
	t := time.NewTicker(d)
	c := make(chan time.Time, 1)
	c <- time.Now() // Il tick immediato

	go func() {
		for tick := range t.C {
			c <- tick
		}
	}()

	return c, t.Stop
}
