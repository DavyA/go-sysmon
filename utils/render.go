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
	var sb strings.Builder

	// ANSI: Move Home and Clear Screen (Atomic Double-Buffering)
	sb.WriteString("\033[H\033[2J")

	// 1. HEADER (Sempre visibile)
	renderHeader(s, cols, &sb)

	if rows < 5 || cols < 30 {
		sb.WriteString("\r\n  Terminal too small!\r\n")
	} else {
		switch s.CurrentView {
		case ViewDashboard:
			renderSideBySide(s, cols, &sb)
		case ViewGPU:
			RenderGPU(s, cols, &sb)
		case ViewPIDs:
			renderPIDs(s, cols, rows, &sb)
		case ViewInfo:
			RenderNeofetch(s, &sb)
		case ViewCPU:
			sb.WriteString(fmt.Sprintf("\r\n  %s[ Advanced CPU View - Under Construction ]%s\r\n", Cyan, Reset))
		}
	}

	renderFooter(s, cols, rows, &sb)

	// Flush the buffer to stdout
	fmt.Print(sb.String())
}

func renderSummaryLine(s *SystemState, cols int, sb *strings.Builder) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	totalPower := s.CPUStats.PackagePower
	for _, g := range s.GPUs { totalPower += g.Power }
	maxTemp := 0.0; for _, t := range s.CPUTemp { if t > maxTemp { maxTemp = t } }
	
	sysPower := totalPower
	if s.Battery.Present && s.Battery.Status == "Discharging" && s.Battery.Power > 0 {
		sysPower = s.Battery.Power
	}

	batStatus := "AC"
	if s.Battery.Present {
		icon := "⚡"
		if s.Battery.Status == "Discharging" { icon = "🔋" }
		batStatus = fmt.Sprintf("%d%% %s", s.Battery.Capacity, icon)
	}

	summary := fmt.Sprintf(" %sSYS PWR: %.1f W%s | %sTEMP: %.1f°C%s | %sBAT: %s%s | %sUP: %s%s",
		Bold, sysPower, Reset, Bold, maxTemp, Reset, Bold, batStatus, Reset, Bold, getUptime(), Reset)

	sb.WriteString(fmt.Sprintf("\r\n  %s\r\n", summary))
	sb.WriteString(fmt.Sprintf("  %s%s%s\r\n", Gray, strings.Repeat("┈", cols-4), Reset))
}

func renderSideBySide(s *SystemState, cols int, sb *strings.Builder) {
	renderSummaryLine(s, cols, sb)

	s.Mu.RLock()
	defer s.Mu.RUnlock() // 1. Definiamo i pesi: W1+W2 = cols - 7
	available := cols - 7
	if available < 10 {
		return
	}

	cpuWidth := (available * 60) / 100
	if cpuWidth > 80 {
		cpuWidth = 80
	}
	memWidth := available - cpuWidth

	cpuLines := []string{}
	memLines := []string{}

	// CPU Cores - 2 per riga
	for i := 0; i < len(s.CPUStats.CpuCores); i += 2 {
		line := ""
		c1 := s.CPUStats.CpuCores[i]
		line += fmt.Sprintf("%-5s [%s] %2.0f%%", c1.ID, CreateSolidBar(c1.Usage), c1.Usage)
		if i+1 < len(s.CPUStats.CpuCores) {
			c2 := s.CPUStats.CpuCores[i+1]
			line += fmt.Sprintf(" | %-5s [%s] %2.0f%%", c2.ID, CreateSolidBar(c2.Usage), c2.Usage)
		}
		cpuLines = append(cpuLines, line)
	}

	// Memoria
	memPercent := (float64(s.Memory.MemUsed) / float64(s.Memory.MemTotal)) * 100
	memLines = append(memLines, fmt.Sprintf("RAM  [%s] %d%%", CreateSolidBar(memPercent), int(memPercent)))
	memLines = append(memLines, fmt.Sprintf("Used: %d / %d MB", s.Memory.MemUsed, s.Memory.MemTotal))
	if s.Memory.SwapTotal > 0 {
		swapPct := (float64(s.Memory.SwapUsed) / float64(s.Memory.SwapTotal)) * 100
		memLines = append(memLines, fmt.Sprintf("SWAP [%s] %d%%", CreateSolidBar(swapPct), int(swapPct)))
	}

	// GPU
	if len(s.GPUs) > 0 {
		memLines = append(memLines, "")
		for _, g := range s.GPUs {
			memLines = append(memLines, fmt.Sprintf("GPU %s: %2.0f%% | %.1fW", g.ID, g.Utilization, g.Power))
		}
	}

	// Network
	rx, tx := 0.0, 0.0
	for _, ni := range s.NetStats {
		rx += ni.RxSpeed
		tx += ni.TxSpeed
	}
	if rx > 0 || tx > 0 {
		memLines = append(memLines, "")
		memLines = append(memLines, fmt.Sprintf("NET RX: %s%.2f MB/s%s", Cyan, rx, Reset))
		memLines = append(memLines, fmt.Sprintf("NET TX: %s%.2f MB/s%s", Cyan, tx, Reset))
	}

	// Disks
	if len(s.Disks) > 0 {
		memLines = append(memLines, "")
		for _, d := range s.Disks {
			if d.MountPoint == "/" || d.MountPoint == "/home" {
				memLines = append(memLines, fmt.Sprintf("DISK %-5s %2.0f%% [%dGB]", d.MountPoint, d.Percent, d.Total/1024))
			}
		}
	}

	maxLines := len(cpuLines)
	if len(memLines) > maxLines {
		maxLines = len(memLines)
	}

	formatTitle := func(title string, width int) string {
		vLen := VisibleLen(title)
		left := (width - vLen) / 2
		right := width - vLen - left
		if left < 0 { left = 0 }
		if right < 0 { right = 0 }
		return strings.Repeat("─", left) + title + strings.Repeat("─", right)
	}

	// Border Top
	sb.WriteString(fmt.Sprintf("%s┌%s┬%s┐%s\r\n", Blue, formatTitle(" CPU CORES ", cpuWidth+2), formatTitle(" SYSTEM ", memWidth+2), Reset))

	if maxLines == 0 {
		// Placeholder row
		sb.WriteString(fmt.Sprintf("%s│ %s", Blue, Reset))
		writePadded(sb, "Collecting metrics...", cpuWidth)
		sb.WriteString(fmt.Sprintf(" %s│ %s", Blue, Reset))
		writePadded(sb, "...", memWidth)
		sb.WriteString(fmt.Sprintf(" %s│%s\r\n", Blue, Reset))
	} else {
		for i := 0; i < maxLines; i++ {
			lLine, rLine := "", ""
			if i < len(cpuLines) { lLine = cpuLines[i] }
			if i < len(memLines) { rLine = memLines[i] }

			sb.WriteString(fmt.Sprintf("%s│ %s", Blue, Reset))
			writePadded(sb, lLine, cpuWidth)
			sb.WriteString(fmt.Sprintf(" %s│ %s", Blue, Reset))
			writePadded(sb, rLine, memWidth)
			sb.WriteString(fmt.Sprintf(" %s│%s\r\n", Blue, Reset))
		}
	}

	// Border Bottom
	sb.WriteString(fmt.Sprintf("%s└%s┴%s┘%s\r\n", Blue, strings.Repeat("─", cpuWidth+2), strings.Repeat("─", memWidth+2), Reset))
}

func writePadded(sb *strings.Builder, text string, width int) {
	vLen := VisibleLen(text)

	if vLen > width && width > 3 {
		sb.WriteString(text)
	} else {
		sb.WriteString(text)
		if width > vLen {
			sb.WriteString(strings.Repeat(" ", width-vLen))
		}
	}
}

func VisibleLen(text string) int {
	visibleLen := 0
	inEsc := false
	for _, r := range text {
		if r == '\033' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		visibleLen++
	}
	return visibleLen
}


func renderFooter(state *SystemState, cols int, rows int, sb *strings.Builder) {
	sb.WriteString(fmt.Sprintf("\033[%d;1H", rows))

	tabs := []string{"[1] Dash", "[2] CPU", "[3] GPU", "[4] PIDs", "[5] Info"}
	footer := ""

	for i, t := range tabs {
		if int(state.CurrentView) == i {
			footer += fmt.Sprintf("\033[44;37m %s \033[0m ", t) 
		} else {
			footer += fmt.Sprintf("\033[90m %s \033[0m ", t) 
		}
	}

	vLen := VisibleLen(footer)
	padding := cols - vLen - 7
	if padding < 0 { padding = 0 }
	sb.WriteString(fmt.Sprintf("%s%*s\033[0m", footer, padding, "Q:Quit"))
}

func renderHeader(s *SystemState, cols int, sb *strings.Builder) {
	s.Mu.RLock()
	defer s.Mu.RUnlock()
	sb.WriteString(fmt.Sprintf("┌%s┐\r\n", strings.Repeat("─", cols-2)))
	header := fmt.Sprintf(" HOST: %s | CPU: %s ", s.Hostname, s.Model)
	if len(header) > cols-4 {
		header = header[:cols-7] + "..." 
	}
	sb.WriteString(fmt.Sprintf("│ %-*s │\r\n", cols-4, header))
	sb.WriteString(fmt.Sprintf("└%s┘\r\n", strings.Repeat("─", cols-2)))
}

func RenderGPU(s *SystemState, cols int, sb *strings.Builder) {
	s.Mu.RLock()
	gpus := s.GPUs
	history := s.GPUHistory
	s.Mu.RUnlock()

	if len(gpus) == 0 {
		sb.WriteString(fmt.Sprintf("\r\n  %sNo GPUs directly detected/supported.%s\r\n", Red, Reset))
	} else {
		for _, g := range gpus {
			sb.WriteString(fmt.Sprintf("\r\n  %s[%s] %s%s\r\n", Cyan, g.ID, g.Name, Reset))

			sb.WriteString(fmt.Sprintf("  Driver: %s | Vendor: %s\r\n", g.Driver, g.Vendor))
			if g.Power > 0 {
				sb.WriteString(fmt.Sprintf("  Power:  %s%.2f W%s  | Temp: %s%.1f°C%s\r\n", Bold, g.Power, Reset, Bold, g.Temp, Reset))
			} else if g.Temp > 0 {
				sb.WriteString(fmt.Sprintf("  Temp:   %s%.1f°C%s\r\n", Bold, g.Temp, Reset))
			}

			barGraph := CreateSolidBar(g.Utilization)
			sb.WriteString(fmt.Sprintf("  Load:   [%s] %3.0f%%\r\n", barGraph, g.Utilization))

			if hist, exists := history[g.ID]; exists && len(hist) > 0 {
				graphWidth := cols - 16
				if graphWidth > 10 {
					tRow, midRow, bRow := CreateSolidGraph(hist, graphWidth)
					sb.WriteString(fmt.Sprintf("          %s\r\n", tRow))
					sb.WriteString(fmt.Sprintf("          %s\r\n", midRow))
					sb.WriteString(fmt.Sprintf("          %s\r\n", bRow))
				}
			}

			if g.MemTotal > 0 {
				memPct := (float64(g.MemUsed) / float64(g.MemTotal)) * 100
				if memPct > 100 { memPct = 100 }
				memBar := CreateSolidBar(memPct)
				sb.WriteString(fmt.Sprintf("  VRAM:   [%s] %d / %d MB\r\n", memBar, g.MemUsed, g.MemTotal))
			}
		}
	}
}
func renderPIDs(s *SystemState, cols int, rows int, sb *strings.Builder) {
	renderSummaryLine(s, cols, sb)

	s.Mu.RLock()
	defer s.Mu.RUnlock()

	// Title centered
	title := " PROCESS LIST (Sort: CPU%) "
	sb.WriteString(fmt.Sprintf("%s%s%s\r\n", Bold+Cyan, title, Reset))

	// Table Header with consistent spacing
	header := fmt.Sprintf(" %-7s %-12s %6s %9s %-5s %s", "PID", "USER", "CPU%", "MEM", "STAT", "COMMAND")
	sb.WriteString(fmt.Sprintf("%s%s%s\r\n", Bold+Blue, header, Reset))
	sb.WriteString(fmt.Sprintf("%s%s%s\r\n", Gray, strings.Repeat("┈", cols-4), Reset))

	// Header(3) + Summary(3) + PID Header(3) + Footer(1) = 10
	startRow := 10
	availableRows := rows - startRow - 2
	if availableRows < 1 {
		return
	}

	procs := s.Processes
	numProcs := len(procs)
	if numProcs == 0 {
		sb.WriteString("\r\n   Collecting process info...\r\n")
		return
	}

	// Clamp cursor
	if s.ProcCursor >= numProcs {
		s.ProcCursor = numProcs - 1
	}
	if s.ProcCursor < 0 {
		s.ProcCursor = 0
	}

	startIdx := 0
	if s.ProcCursor >= availableRows {
		startIdx = s.ProcCursor - availableRows + 1
	}

	endIdx := startIdx + availableRows
	if endIdx > numProcs {
		endIdx = numProcs
	}

	for i := startIdx; i < endIdx; i++ {
		p := procs[i]
		pidStr := fmt.Sprintf("%-7d", p.PID)
		userStr := fmt.Sprintf("%-12s", p.User)
		if len(userStr) > 12 {
			userStr = userStr[:11] + "…"
		}

		cpuColor := Reset
		if p.CPU > 50 {
			cpuColor = Red
		} else if p.CPU > 10 {
			cpuColor = Yellow
		}
		cpuStr := fmt.Sprintf("%s%5.1f%%%s", cpuColor, p.CPU, Reset)
		
		memStr := fmt.Sprintf("%7d MB", p.MemRSS)
		statStr := fmt.Sprintf("%-5s", p.Status)
		gpuStr := "   "
		if p.IsGPU {
			gpuStr = Bold + Green + "[G]" + Reset
		}
		cmdStr := p.CmdLine

		line := fmt.Sprintf(" %s %s %s %s %s %s %s", pidStr, userStr, cpuStr, memStr, statStr, gpuStr, cmdStr)

		if i == s.ProcCursor {
			// Strippiamo i colori per calcolare la lunghezza visibile nel highlight
			vLinePlain := fmt.Sprintf(" %s %s %5.1f%% %s %s %s %s", pidStr, userStr, p.CPU, memStr, statStr, strings.Trim(gpuStr, Bold+Green+Reset), cmdStr)
			if VisibleLen(vLinePlain) > cols-1 {
				vLinePlain = vLinePlain[:cols-4] + "..."
			}
			sb.WriteString(fmt.Sprintf("\033[44;37m%-*s\033[0m\r\n", cols, vLinePlain))
		} else {
			if VisibleLen(line) > cols-1 {
				line = line[:cols-4] + "..."
			}
			sb.WriteString(line + "\r\n")
		}
	}
}
