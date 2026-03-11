package utils

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// Colours for the CLI
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
	Bold   = "\033[1m"
)

var blocks = []string{"▂", "▃", "▄", "▅", "▆", "▇", "█"}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// get the block character for the graph
func getBlockChar(percent float64) string {
	// Mapping 0-100% su 8 livelli
	idx := int(percent * float64(len(blocks)-1) / 100)
	if idx > len(blocks)-1 {
		idx = len(blocks) - 1
	}
	if idx < 0 {
		idx = 0
	}
	return blocks[idx]
}

// CreateSolidGraph creates a solid graph of the CPU usage
// history is a slice of float64 representing the CPU usage percentage(0-100)
// width is the width of the graph
func CreateSolidGraph(history []float64, width int) (string, string, string) {
	if len(history) < 1 {
		return "", "", ""
	}

	// Prendiamo gli ultimi 'width' campioni (1 campione = 1 carattere di larghezza)
	start := len(history) - width
	if start < 0 {
		start = 0
	}
	samples := history[start:]

	var topRow, midRow, bottomRow strings.Builder

	for _, val := range samples {
		var tChar, midChar, bChar string
		color := "\033[32m" // Verde
		if val > 50 {
			color = "\033[33m"
		} // Giallo
		if val > 80 {
			color = "\033[31m"
		} // Rosso

		// RIGA SOTTO (0-33%)
		if val <= 33 {
			// Mapping 0-33 su 0-8 livelli dei blocchi
			bChar = getBlockChar(val * 3)
			midChar = " "
			tChar = " " // Vuoto sopra
		} else if val <= 66 {
			// Se sopra il 33%, la riga sotto è PIENA
			bChar = "█"
			// RIGA SOPRA (34-66%)
			midChar = getBlockChar((val - 33) * 3)
			tChar = " "
		} else {
			// Se sopra il 66%, la riga sotto è PIENA
			bChar = "█"
			midChar = "█"
			// RIGA SOPRA (67-100%)
			tChar = getBlockChar((val - 66) * 3)
		}

		bottomRow.WriteString(color + bChar + "\033[0m")
		midRow.WriteString(color + midChar + "\033[0m")
		topRow.WriteString(color + tChar + "\033[0m")
	}
	return topRow.String(), midRow.String(), bottomRow.String()
}

// CreateSolidBar creates a solid bar of the CPU usage
// value is a float64 representing the CPU usage percentage(0-100)
func CreateSolidBar(value float64) string {

	var bar strings.Builder

	color := "\033[32m" // Verde
	if value > 50 {
		color = "\033[33m"
	} // Giallo
	if value > 80 {
		color = "\033[31m"
	} // Rosso

	bar.WriteString(color + getBlockChar(value) + "\033[0m")

	return bar.String()
}

// GetTermSize gets the size of the terminal
func GetTermSize() (int, int) {
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 80, 24 // Fallback standard
	}
	return int(ws.Col), int(ws.Row)
}

func HandleInput(state *SystemState, RedrawChan chan bool) {
	oldState, _ := SetRawMode(os.Stdin.Fd())
	defer RestoreMode(os.Stdin.Fd(), oldState)
	buf := make([]byte, 3)
	for {
		n, _ := os.Stdin.Read(buf)
		if n > 0 {
			state.Mu.Lock()
			switch buf[0] {
			case '1':
				state.CurrentView = ViewDashboard
				RedrawChan <- true
			case '2':
				state.CurrentView = ViewCPU
				RedrawChan <- true
			case '3':
				state.CurrentView = ViewGPU
				RedrawChan <- true
			case '4':
				state.CurrentView = ViewPIDs
				RedrawChan <- true
			case '5':
				state.CurrentView = ViewInfo
				RedrawChan <- true
			case 'q', 3: // 'q' o CTRL+C
				fmt.Print("\033[?25h\033[?1049l")
				RestoreMode(os.Stdin.Fd(), oldState)
				os.Exit(0)
			}
			state.Mu.Unlock()
		}
	}
}

func SetRawMode(fd uintptr) (*syscall.Termios, error) {
	var old syscall.Termios
	// Usa TCGETS (0x5401 su molti sistemi Linux x86_64)
	// Se TCGETS non è definito, usa il valore numerico diretto 0x5401
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(0x5401), uintptr(unsafe.Pointer(&old))); err != 0 {
		return nil, err
	}

	newTerm := old
	// Lflag: Disabilita Echo, Modalità Canonica (attesa Invio), e Segnali (Ctrl+C)
	// ISIG lo togliamo se vuoi gestire Ctrl+C manualmente come 'q'
	newTerm.Lflag &^= (syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN)

	// Iflag: Disabilita break, conversioni CR/NL e controllo di flusso (Ctrl+S/Q)
	newTerm.Iflag &^= (syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON)

	// Oflag: Disabilita il post-processing dell'output
	newTerm.Oflag &^= (syscall.OPOST)

	// VMIN e VTIME: Rendi la Read non bloccante o bloccante per 1 singolo carattere
	newTerm.Cc[syscall.VMIN] = 1  // Leggi almeno 1 byte
	newTerm.Cc[syscall.VTIME] = 0 // Nessun timeout

	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(0x5402), uintptr(unsafe.Pointer(&newTerm))); err != 0 {
		return nil, err
	}

	return &old, nil
}

func RestoreMode(fd uintptr, old *syscall.Termios) {
	syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCSETS, uintptr(unsafe.Pointer(old)), 0, 0, 0)
}
