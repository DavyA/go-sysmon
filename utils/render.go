package utils

import (
	"fmt"
	"os"
	"strings"
)

func HandleResize(sigChan chan os.Signal, redrawChan chan bool) {
	for range sigChan {
		redrawChan <- true // Notifica al main di ridisegnare subito
	}
}

func (s *SystemState) Render() {
	cols, rows := GetTermSize()
	// 1. HEADER (Sempre visibile)
	renderHeader(s, cols)

	switch s.CurrentView {
	case ViewDashboard:
		renderSideBySide(s, cols)
		// ... aggiungi grafico ...
	case ViewGPU:
		RenderGPU(s, cols)
	case ViewPIDs:
		fmt.Printf("\r\n  %s[ Processes View - Under Construction ]%s\r\n", Cyan, Reset)
	case ViewInfo:
		RenderNeofetch(s)
	case ViewCPU:
		fmt.Printf("\r\n  %s[ Advanced CPU View - Under Construction ]%s\r\n", Cyan, Reset)
		//renderFullCPUView(s, cols) // Una vista con grafici per ogni core!
	}

	renderFooter(s, cols, rows)

	/*
			// 2. LOGICA ADATTIVA: Affiancare o Incolonnare?
			// Se la finestra è larga > 100 colonne, mettiamo CPU e RAM affiancate
			if cols > 100 {
				renderSideBySide(s, cols)
			} else {
				// TODO: Implementare il layout stacked
				fmt.Println("Finestra troppo stretta per il layout avanzato.")
				//renderStacked(s, cols)
			}

		// 3. IL GRAFICO (Larghezza dinamica)
		// Usiamo tutta la larghezza disponibile per il grafico
		graphWidth := cols - 62
		tRow, midRow, bRow := CreateSolidGraph(s.CPUHistory, graphWidth)

		if rows > 15 { // Mostra il grafico solo se c'è abbastanza spazio verticale
			fmt.Printf("\n [ CPU HISTORY (W:%d) ]\n", graphWidth)
			fmt.Printf("  %s\n  %s\n  %s\n", tRow, midRow, bRow)
		}
	*/
}

func renderSideBySide(s *SystemState, cols int) {
	// Definiamo i pesi: 60% spazio alla CPU, 40% alla RAM
	cpuWidth := (cols * 60) / 100
	memWidth := cols - cpuWidth - 4 // -4 per i bordi e separatori

	// Prepariamo le "colonne" di testo
	cpuLines := []string{}
	memLines := []string{}

	// 1. Popoliamo la colonna CPU

	// Ricerca la temp della CPU principale (Tctl / Package)
	maxTemp := 0.0
	for k, t := range s.CPUTemp {
		if strings.Contains(strings.ToLower(k), "tctl") || strings.Contains(strings.ToLower(k), "package") || t > maxTemp {
			maxTemp = t
		}
	}

	// Riga riassuntiva di Package
	summaryLine := ""
	if maxTemp > 0 {
		summaryLine += fmt.Sprintf(" Package Temp: %.1f °C ", maxTemp)
	}
	if s.CPUStats.PackagePower > 0 {
		summaryLine += fmt.Sprintf("|  Power: %.2f W", s.CPUStats.PackagePower)
	}
	if summaryLine != "" {
		cpuLines = append(cpuLines, summaryLine)
		cpuLines = append(cpuLines, "") // separator
	}

	// Aggiungiamo i core (2 per riga)
	for i := 0; i < len(s.CPUStats.CpuCores); i += 2 {
		line := ""
		// Core A
		c1 := s.CPUStats.CpuCores[i]
		bar1 := CreateSolidBar(c1.Usage) // bar di 10 char
		line += fmt.Sprintf("%-5s [%s] %3.0f%% ", c1.ID, bar1, c1.Usage)

		// Core B (se esiste)
		if i+1 < len(s.CPUStats.CpuCores) {
			c2 := s.CPUStats.CpuCores[i+1]
			bar2 := CreateSolidBar(c2.Usage)
			line += fmt.Sprintf(" | %-5s [%s] %3.0f%%", c2.ID, bar2, c2.Usage)
		}
		cpuLines = append(cpuLines, line)
	}

	// 2. Popoliamo la colonna RAM/SWAP
	memPercent := (float64(s.Memory.MemUsed) / float64(s.Memory.MemTotal)) * 100
	memLines = append(memLines, fmt.Sprintf("RAM  [%s] %v%%", CreateSolidBar(memPercent), int(memPercent)))
	memLines = append(memLines, fmt.Sprintf("Used: %v / %v MB", s.Memory.MemUsed, s.Memory.MemTotal))

	if s.Memory.SwapTotal > 0 {
		swapPercent := (float64(s.Memory.SwapUsed) / float64(s.Memory.SwapTotal)) * 100
		memLines = append(memLines, "") // Spazio vuoto
		memLines = append(memLines, fmt.Sprintf("SWAP [%s] %v%%", CreateSolidBar(swapPercent), int(swapPercent)))
	}

	// 3. MERGE delle colonne (La parte difficile)
	// Troviamo chi ha più righe per non tagliare nulla
	maxLines := len(cpuLines)
	if len(memLines) > maxLines {
		maxLines = len(memLines)
	}

	cpuTitle := "CPU CORES"
	memTitle := "MEMORY"

	cpuPadding := (cpuWidth - len(cpuTitle)) / 2
	memPadding := (memWidth - len(memTitle)) / 2

	fmt.Printf("%s┌%s%s%s┬%s%s%s┐%s\r\n", Blue, strings.Repeat("─", cpuPadding), cpuTitle, strings.Repeat("─", cpuPadding), strings.Repeat("─", memPadding), memTitle, strings.Repeat("─", memPadding), Reset)

	for i := 0; i < maxLines; i++ {
		left := ""
		if i < len(cpuLines) {
			left = cpuLines[i]
		}

		right := ""
		if i < len(memLines) {
			right = memLines[i]
		}

		// Stampiamo le due parti con padding preciso
		fmt.Printf("%s│ %s%-*s %s│ %s%-*s %s│%s\r\n",
			Blue, Reset, cpuWidth, left, Blue, Reset, memWidth, right, Blue, Reset)
	}
	fmt.Printf("%s└%s┴%s┘%s\r\n", Blue, strings.Repeat("─", cpuWidth), strings.Repeat("─", memWidth), Reset)
}

func renderFooter(state *SystemState, cols int, rows int) {
	// Ci spostiamo all'ultima riga
	fmt.Printf("\033[%d;1H", rows)

	tabs := []string{"[1] Dash", "[2] CPU", "[3] GPU", "[4] PIDs", "[5] Info"}
	footer := ""

	for i, t := range tabs {
		if int(state.CurrentView) == i {
			footer += fmt.Sprintf("\033[44;37m %s \033[0m ", t) // Blu per attiva
		} else {
			footer += fmt.Sprintf("\033[90m %s \033[0m ", t) // Grigio per inattiva
		}
	}

	// Riempiamo il resto della riga con spazio nero
	padding := cols - len(tabs)*11 // calcolo approssimativo
	fmt.Printf("%s%*s\033[0m", footer, padding, "Q:Quit")
}

func renderHeader(s *SystemState, cols int) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	fmt.Printf("┌%s┐\r\n", strings.Repeat("─", cols-2))
	header := fmt.Sprintf(" HOST: %s | CPU: %s ", s.Hostname, s.Model)
	if len(header) > cols-4 {
		header = header[:cols-7] + "..." // Taglia se la finestra è stretta
	}
	fmt.Printf("│ %-*s │\r\n", cols-4, header)
	fmt.Printf("└%s┘\r\n", strings.Repeat("─", cols-2))
}

func RenderGPU(s *SystemState, cols int) {
	s.Mu.RLock()
	gpus := s.GPUs
	history := s.GPUHistory
	s.Mu.RUnlock()

	if len(gpus) == 0 {
		fmt.Printf("\r\n  %sNo GPUs directly detected/supported.%s\r\n", Red, Reset)
	} else {
		for _, g := range gpus {
			fmt.Printf("\r\n  %s[%s] %s%s\r\n", Cyan, g.ID, g.Name, Reset)

			// Statistiche principali
			fmt.Printf("  Driver: %s | Vendor: %s\r\n", g.Driver, g.Vendor)
			if g.Power > 0 {
				fmt.Printf("  Power:  %s%.2f W%s  | Temp: %s%.1f°C%s\r\n", Bold, g.Power, Reset, Bold, g.Temp, Reset)
			} else if g.Temp > 0 {
				fmt.Printf("  Temp:   %s%.1f°C%s\r\n", Bold, g.Temp, Reset)
			}

			// Usage Bar
			barGraph := CreateSolidBar(g.Utilization)
			fmt.Printf("  Load:   [%s] %3.0f%%\r\n", barGraph, g.Utilization)

			// Graph History (if initialized)
			if hist, exists := history[g.ID]; exists && len(hist) > 0 {
				graphWidth := cols - 16
				if graphWidth > 10 {
					tRow, midRow, bRow := CreateSolidGraph(hist, graphWidth)
					fmt.Printf("          %s\r\n", tRow)
					fmt.Printf("          %s\r\n", midRow)
					fmt.Printf("          %s\r\n", bRow)
				}
			}

			// VRAM Bar
			if g.MemTotal > 0 {
				memPct := (float64(g.MemUsed) / float64(g.MemTotal)) * 100
				if memPct > 100 {
					memPct = 100
				}
				memBar := CreateSolidBar(memPct)
				fmt.Printf("  VRAM:   [%s] %d / %d MB\r\n", memBar, g.MemUsed, g.MemTotal)
			}
		}
	}
}
