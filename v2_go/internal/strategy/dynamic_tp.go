package strategy

import (
	"legendary-scalper/internal/config"
)

// CalculateDynamicTarget calculates the Take Profit price based on step count
// Heavy bags (higher steps) require larger profit to justify the risk.
func CalculateDynamicTarget(avgEntry float64, step int, cfg *config.StrategyConfig) float64 {
	baseTP := cfg.Martingale.TakeProfit

	// Dynamic Scaling
	// Steps 1-3: Base TP (1.5%)
	// Steps 4-7: 1.2x (1.8%)
	// Steps 8+:  1.5x (2.25%) - Big reward for big risk

	multiplier := 1.0
	if step >= 8 {
		multiplier = 1.5
	} else if step >= 4 {
		multiplier = 1.2
	}

	targetPercent := baseTP * multiplier

	// Calculate Price (Short = Lower target)
	// target = entry * (1 - percent)
	targetPrice := avgEntry * (1 - (targetPercent / 100.0))

	return targetPrice
}
