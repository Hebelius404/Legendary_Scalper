package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
)

var (
	// Colors using fatih/color for cross-platform support (handles Windows mostly)
	Green   = color.New(color.FgGreen).SprintfFunc()
	Red     = color.New(color.FgRed).SprintfFunc()
	Yellow  = color.New(color.FgYellow).SprintfFunc()
	Cyan    = color.New(color.FgCyan).SprintfFunc()
	White   = color.New(color.FgWhite).SprintfFunc()
	Blue    = color.New(color.FgBlue).SprintfFunc()
	Magenta = color.New(color.FgMagenta).SprintfFunc()

	BoldGreen = color.New(color.FgGreen, color.Bold).SprintfFunc()
	BoldRed   = color.New(color.FgRed, color.Bold).SprintfFunc()
	BoldCyan  = color.New(color.FgCyan, color.Bold).SprintfFunc()
)

// ConsoleUI handles all user visible output
type ConsoleUI struct {
	UseTestnet bool
}

func NewConsoleUI(useTestnet bool) *ConsoleUI {
	return &ConsoleUI{UseTestnet: useTestnet}
}

// PrintBanner displays the startup banner
func (ui *ConsoleUI) PrintBanner(pairsCount, interval int) {
	fmt.Println(Cyan("╔══════════════════════════════════════════════════════════════╗"))
	fmt.Printf("%s  %s%s%s                  %s\n", Cyan("║"), Yellow("🚀 LEGENDARY SCALPER v2.0 (GO EDITION) 🚀"), "", Cyan(""), Cyan("║"))
	fmt.Println(Cyan("║                                                              ║"))
	fmt.Printf("%s  %sStrat: Martingale + RSI + Eagle Eye%s                        %s\n", Cyan("║"), White(""), Cyan(""), Cyan("║"))
	fmt.Printf("%s  %sPairs: %d pairs every %ds%s                                  %s\n", Cyan("║"), White(""), pairsCount, interval, Cyan(""), Cyan("║"))

	modeColor := Red
	modeText := "PRODUCTION"
	if ui.UseTestnet {
		modeColor = Green
		modeText = "TESTNET   "
	}
	fmt.Printf("%s  %sMode: %s%s                                         %s\n", Cyan("║"), White(""), modeColor(modeText), Cyan(""), Cyan("║"))
	fmt.Println(Cyan("╚══════════════════════════════════════════════════════════════╝"))
	fmt.Println()
}

// PrintScanHeader prints the clear separator for each scan loop
func (ui *ConsoleUI) PrintScanHeader(scanNum int, pairsCount int) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Println("\n" + Cyan("============================================================"))
	fmt.Printf("%s 📊 Scan #%d | %s | Analyzing %d pairs %s\n", Cyan(""), scanNum, timestamp, pairsCount, Cyan(""))
	fmt.Println(Cyan("============================================================"))
}

// LogInfo prints a standard info message
func (ui *ConsoleUI) LogInfo(msg string) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("%s | %s | %s\n", ts, Green("INFO "), msg)
}

// LogError prints an error message
func (ui *ConsoleUI) LogError(msg string) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("%s | %s | %s\n", ts, Red("ERROR"), msg)
}

// LogWarning prints a warning message
func (ui *ConsoleUI) LogWarning(msg string) {
	ts := time.Now().Format("15:04:05")
	fmt.Printf("%s | %s | %s\n", ts, Yellow("WARN "), msg)
}

// LogSignal prints a detected trade signal
func (ui *ConsoleUI) LogSignal(symbol, side string, price float64) {
	ts := time.Now().Format("15:04:05")
	icon := "🟢"
	colorFunc := BoldGreen

	if side == "SELL" { // For Shorting
		icon = "🔴"
		colorFunc = BoldRed
	}

	fmt.Printf("%s | %s %s SIGNAL: %s %s | Price: %.4f\n",
		ts, icon, colorFunc(side), colorFunc(symbol), White(""), price)
}

// LogPump prints a pump detection line
func (ui *ConsoleUI) LogPump(symbol string, change, volume float64, isTop bool) {
	ts := time.Now().Format("15:04:05")
	icon := "🔍"
	if isTop {
		icon = "🚀"
	}

	fmt.Printf("%s | %s %s: %s (%.2f%%) Vol: $%.0f\n",
		ts, icon, Cyan("Pump"), BoldCyan(symbol), change, volume)
}

// DisplayPositions prints the table of active trades
func (ui *ConsoleUI) DisplayPositions(positions []string) {
	if len(positions) == 0 {
		// Only print "No positions" periodically or not at all to reduce noise?
		// Python prints "📭 No open positions"
		// Let's implement that method if needed, but for now typical logs.
		return
	}

	fmt.Printf("\n%s 📈 Open Positions (%d):\n", White(""), len(positions))
	for _, pos := range positions {
		fmt.Println("  " + pos)
	}
}
