package bot

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"legendary-scalper/internal/ai"
	"legendary-scalper/internal/config"
	"legendary-scalper/internal/exchange"
	"legendary-scalper/internal/scanner"
	"legendary-scalper/internal/strategy"
	"legendary-scalper/internal/ui"
)

// BotEngine orchestrates the entire trading system
type BotEngine struct {
	Cfg       *config.AppConfig
	Client    exchange.ExchangeClient
	Scanner   *scanner.ScannerEngine
	Strategy  *strategy.MartingaleStrategy
	AI        *ai.GrokClient
	State     *strategy.StateManager
	UI        *ui.ConsoleUI
	Running   bool
	ScanCount int
}

// NewBotEngine initializes all dependencies
func NewBotEngine(cfg *config.AppConfig) *BotEngine {
	// 1. Initialize UI
	console := ui.NewConsoleUI(cfg.Binance.UseTestnet)

	// 2. Initialize Exchange
	client := exchange.NewBinanceClient(
		cfg.Binance.APIKey,
		cfg.Binance.APISecret,
		cfg.Binance.UseTestnet,
	)

	// 3. Initialize Components
	sc := scanner.NewScannerEngine(&cfg.Scanning, client)
	st := strategy.NewMartingaleStrategy(&cfg.Strategy, client)
	aiClient := ai.NewGrokClient(cfg)
	state := strategy.NewStateManager("positions.json")

	// Load previous state
	if err := state.LoadState(); err != nil {
		console.LogWarning(fmt.Sprintf("Failed to load state: %v", err))
	} else {
		console.LogInfo(fmt.Sprintf("Loaded %d active positions from disk", len(state.ActiveTrades)))
	}

	return &BotEngine{
		Cfg:       cfg,
		Client:    client,
		Scanner:   sc,
		Strategy:  st,
		AI:        aiClient,
		State:     state,
		UI:        console,
		Running:   true,
		ScanCount: 0,
	}
}

// Run starts the main infinite loop
func (b *BotEngine) Run() {
	b.UI.PrintBanner(b.Cfg.Scanning.TopPairs, b.Cfg.Scanning.Interval)

	ticker := time.NewTicker(time.Duration(b.Cfg.Scanning.Interval) * time.Second)
	defer ticker.Stop()

	// Initial Scan
	b.Tick(context.Background())

	for range ticker.C {
		if !b.Running {
			break
		}
		b.Tick(context.Background())
	}
}

// Tick executes one logic cycle
func (b *BotEngine) Tick(ctx context.Context) {
	b.ScanCount++
	b.UI.PrintScanHeader(b.ScanCount, b.Cfg.Scanning.TopPairs)

	// 1. Check Positions (Manage existing trades)
	// SYNC: Fetch ALL open positions from Binance to ensure full visibility
	allPos, err := b.Client.GetAllOpenPositions(ctx)
	if err != nil {
		b.UI.LogError(fmt.Sprintf("Failed to sync positions: %v", err))
	} else {
		// Adopt/Update State
		activeSymbols := make(map[string]bool)
		for _, p := range allPos {
			activeSymbols[p.Symbol] = true
			amt, _ := strconv.ParseFloat(p.PositionAmt, 64)
			entryPrice, _ := strconv.ParseFloat(p.EntryPrice, 64)
			leverage, _ := strconv.ParseFloat(p.Leverage, 64)
			if leverage == 0 {
				leverage = 1
			}

			// Check if we track this
			if localPos, exists := b.State.ActiveTrades[p.Symbol]; exists {
				// Update existing
				if localPos.EntryPrice != entryPrice {
					localPos.EntryPrice = entryPrice
					b.State.UpdatePosition(localPos)
				}

				// Sync Quantity and re-infer Step if needed
				// Presumed Margin
				margin := (math.Abs(amt) * entryPrice) / leverage

				// Infer Step from REAL Margin on Exchange
				inferredStep := b.Strategy.InferStepFromMargin(margin)

				if localPos.Step != inferredStep {
					// b.UI.LogInfo(fmt.Sprintf("🔄 Correcting Step for %s: %d -> %d (Margin: $%.2f)", p.Symbol, localPos.Step, inferredStep, margin))
					localPos.Step = inferredStep
					b.State.UpdatePosition(localPos)
				}

				if math.Abs(localPos.Quantity-math.Abs(amt)) > 0.0001 {
					localPos.Quantity = math.Abs(amt)
					b.State.UpdatePosition(localPos)
				}
			} else {
				// Adopt new position found on exchange
				// Presumed Margin (Notional / Leverage)
				margin := (math.Abs(amt) * entryPrice) / leverage

				// Infer Step from Size
				step := b.Strategy.InferStepFromMargin(margin)

				newPos := &strategy.MartingalePosition{
					Symbol:      p.Symbol,
					Step:        step,
					EntryPrice:  entryPrice,
					Quantity:    math.Abs(amt),
					TotalMargin: margin,
					LastAddTime: time.Now(),
				}
				b.State.UpdatePosition(newPos)
				b.UI.LogInfo(fmt.Sprintf("🆕 Adopted active position from Exchange: %s (Step %d)", p.Symbol, step))
				// Refresh allPos immediately or let next tick handle full sync?
				// Engine cycle is fast, next tick is fine.
				// But we are in the loop, so local 'newPos' is now in State.
			}
		}

		// Cleanup: Remove positions from State that are closed on Exchange
		for sym := range b.State.ActiveTrades {
			if !activeSymbols[sym] {
				// It's closed on exchange but open in state -> Remove it
				b.State.RemovePosition(sym)
			}
		}
	}

	activePositions := []string{}
	if len(b.State.ActiveTrades) > 0 {
		for _, p := range b.State.ActiveTrades {
			// Get Mark Price for PnL
			price, _ := b.Client.GetPrice(ctx, p.Symbol)

			pnl := 0.0
			pnlPercent := 0.0

			// Recalculate margin for display approx
			// Or stick to what we saved. If we adopted, we saved calculated margin.
			margin := p.TotalMargin

			if price > 0 {
				// Short PnL: (Entry - Current) * Qty
				// Assumption: User is running Short Scalper per context
				pnl = (p.EntryPrice - price) * p.Quantity
				pnlPercent = (pnl / p.TotalMargin) * 100
			}

			// Format: SYMBOL: Step X/9 | Margin: $XX.XX | PnL: $X.XX (X.XX%)
			posStr := fmt.Sprintf("%s: Step %d/9 | Margin: $%.2f | PnL: $%.2f (%.2f%%)",
				p.Symbol, p.Step, margin, pnl, pnlPercent)

			if pnl >= 0 {
				posStr = fmt.Sprintf("%s: Step %d/9 | Margin: $%.2f | PnL: %s $%.2f (%.2f%%)",
					p.Symbol, p.Step, margin, ui.Green(""), pnl, pnlPercent)
			} else {
				posStr = fmt.Sprintf("%s: Step %d/9 | Margin: $%.2f | PnL: %s $%.2f (%.2f%%)",
					p.Symbol, p.Step, margin, ui.Red(""), pnl, pnlPercent)
			}
			activePositions = append(activePositions, posStr)
		}
	}
	b.UI.DisplayPositions(activePositions)

	// 2. Scan for New Opportunities
	b.UI.LogInfo("Scanning for pumped coins...")
	opps, err := b.Scanner.ScanForOpportunities(ctx)
	if err != nil {
		b.UI.LogError(fmt.Sprintf("Scan failed: %v", err))
		return
	}

	if len(opps) == 0 {
		return
	}

	// Log top opportunities
	best := opps[0]
	b.UI.LogPump(best.Symbol, best.Change24h, best.Volume24h, true)

	// 3. Evaluate Entry (Strategy + AI)

	// A. Check if already in position
	if _, exists := b.State.ActiveTrades[best.Symbol]; exists {
		return // Already trading this
	}

	// B. Check RSI (Safety)
	price, err := b.Client.GetPrice(ctx, best.Symbol)
	if err != nil {
		b.UI.LogWarning(fmt.Sprintf("Failed to get price for %s: %v", best.Symbol, err))
		return
	}

	// Fetch recent Klines for RSI calculation
	klines, err := b.Client.GetKlines(ctx, best.Symbol, "1m", 20)
	if err != nil {
		b.UI.LogWarning(fmt.Sprintf("Failed to get klines: %v", err))
		return
	}

	// Convert klines to prices slice
	var closes []float64
	for _, k := range klines {
		closes = append(closes, k.Close)
	}

	// 4. Calculate Size (Step 1)
	marginUSDT := b.Strategy.GetStepSize(0)
	leverage := float64(b.Cfg.Strategy.Martingale.Leverage)
	if leverage < 1 {
		leverage = 1 // Safety default
	}

	totalNotional := marginUSDT * leverage
	quantity := totalNotional / price

	// Place Order
	b.UI.LogSignal(best.Symbol, "SELL", price)

	// EXECUTE TRADING
	avgPrice, err := b.Client.PlaceEntryOrder(ctx, best.Symbol, "SELL", quantity)
	if err != nil {
		b.UI.LogError(fmt.Sprintf("Execution Failed: %v", err))
		return
	}

	// Fallback: If API returns 0 (market order not fully filled in response), use ticker price
	if avgPrice == 0 {
		avgPrice = price
		b.UI.LogWarning(fmt.Sprintf("API returned 0 fill price for %s, using ticker price: %.4f", best.Symbol, avgPrice))
	}

	// 5. Record State
	newPos := &strategy.MartingalePosition{
		Symbol:      best.Symbol,
		Step:        1,
		EntryPrice:  avgPrice,
		Quantity:    quantity,
		TotalMargin: marginUSDT,
		LastAddTime: time.Now(),
	}
	b.State.UpdatePosition(newPos)
	b.UI.LogInfo(fmt.Sprintf("Position Opened: %s", best.Symbol))
}
