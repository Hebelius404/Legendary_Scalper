package strategy

import (
	"fmt"
)

// CheckSafetyConditions verifies all safety checks before finding a trade
func (s *MartingaleStrategy) CheckSafetyConditions(rsi float64, step int) (bool, string) {
	// 1. RSI Circuit Breaker (Steps 4+)
	// If we are deep in the hole (Step 4+) and RSI is STILL excessively high (>90),
	// it means the pump is parabolic. Adding more now is suicide. Wait for cooling.
	if step >= 4 {
		limit := s.cfg.Safety.RSICircuitBreaker
		if rsi > limit {
			return false, fmt.Sprintf("⛔ RSI Circuit Breaker! RSI %.1f > %.1f. Waiting for cool-off.", rsi, limit)
		}
	}

	return true, "Safety checks passed"
}
