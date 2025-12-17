package strategy

import (
	"fmt"
	"math"
	"time"

	"legendary-scalper/internal/config"
	"legendary-scalper/internal/exchange"
)

// MartingalePosition represents a single active trading sequence
type MartingalePosition struct {
	Symbol        string    `json:"symbol"`
	Step          int       `json:"step"`            // Current step (1-9)
	EntryPrice    float64   `json:"entry_price"`     // Weighted average entry
	Quantity      float64   `json:"quantity"`        // Total position size
	TotalMargin   float64   `json:"total_margin"`    // Total USDT margin used
	NextStepPrice float64   `json:"next_step_price"` // Trigger price for next add
	StopLoss      float64   `json:"stop_loss"`       // Hard stop price
	TakeProfit    float64   `json:"take_profit"`     // Target exit price
	LastAddTime   time.Time `json:"last_add_time"`   // Time of last step add
}

// MartingaleStrategy manages the logic for entering and managing positions
type MartingaleStrategy struct {
	cfg    *config.StrategyConfig
	client exchange.ExchangeClient
}

// NewMartingaleStrategy creates a new manager
func NewMartingaleStrategy(cfg *config.StrategyConfig, client exchange.ExchangeClient) *MartingaleStrategy {
	return &MartingaleStrategy{
		cfg:    cfg,
		client: client,
	}
}

// CalculateNextStep determines the parameters for the next Martingale step
func (s *MartingaleStrategy) CalculateNextStep(pos *MartingalePosition, currentPrice float64, volatility float64) (shouldAdd bool, reason string) {
	// 1. Check Max Steps
	if pos.Step >= len(s.cfg.Martingale.Steps) {
		return false, "Max steps reached"
	}

	// 2. Check Wait Time
	waitTime := time.Duration(s.cfg.Martingale.StepWaitTimes[pos.Step]) * time.Minute
	if time.Since(pos.LastAddTime) < waitTime {
		return false, fmt.Sprintf("Waiting for cooldown (%v remaining)", waitTime-time.Since(pos.LastAddTime))
	}

	// 3. Dynamic Spacing Logic
	// Calculate required drop percentage based on config distance
	baseDistance := s.cfg.Martingale.StepDistances[pos.Step]

	// Apply Volatility Multiplier (Safety Feature)
	// If volatility is high, we widen the gap
	multiplier := 1.0
	if volatility > 0 {
		// Logic: volatility 1.0 = 1x, 2.0 = 1.5x, etc. (Simplified for now)
		multiplier = math.Min(1.0+(volatility*0.5), s.cfg.Safety.MaxVolMultiplier)
	}

	requiredDropPercent := baseDistance * multiplier

	// Check if price has dropped enough
	// For SHORT position: Price should be HIGHER by X%
	// (Assuming we are SHORTing pumps)
	thresholdPrice := pos.EntryPrice * (1 + (requiredDropPercent / 100))

	if currentPrice >= thresholdPrice {
		return true, fmt.Sprintf("Price target hit (Dist: %.2f%%, VolMult: %.1fx)", requiredDropPercent, multiplier)
	}

	return false, "Price target not reached"
}

// GetStepSize returns the margin amount for a specific step
func (s *MartingaleStrategy) GetStepSize(stepIndex int) float64 {
	if stepIndex < 0 || stepIndex >= len(s.cfg.Martingale.Steps) {
		return 0
	}
	return float64(s.cfg.Martingale.Steps[stepIndex])
}

// InferStepFromMargin guesses which step a position is at based on its margin size
func (s *MartingaleStrategy) InferStepFromMargin(marginUSDT float64) int {
	cumulative := 0.0
	// 5% tolerance buffer for fee discrepancies or price shifts
	tolerance := 0.95

	for i, stepSize := range s.cfg.Martingale.Steps {
		cumulative += float64(stepSize)
		// If current margin is roughly equal or less than cumulative, we are at this step
		// (Assuming we filled up to this point)
		// However, since we might be slightly over due to price fluctuation on "EntryPrice",
		// we should find the closest Cumulative that matches.

		// Let's check if the margin fits within this step's accumulated total
		if marginUSDT <= cumulative/tolerance {
			return i + 1 // 1-based index
		}
	}
	// If larger than max configured, it's at max step (or user added manually beyond)
	return len(s.cfg.Martingale.Steps)
}
