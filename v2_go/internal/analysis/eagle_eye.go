package analysis

import (
	"math"
)

// Candle represents a single price bar
type Candle struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

// AnalyzeCandlePatterns checks for bearish reversal signals ("Eagle Eye")
// Returns: match (bool), patternName (string), strength (int 1-5)
func AnalyzeCandlePatterns(candles []Candle) (bool, string, int) {
	if len(candles) < 3 {
		return false, "", 0
	}

	last := candles[len(candles)-1]

	body := math.Abs(last.Close - last.Open)
	upperWick := last.High - math.Max(last.Close, last.Open)
	totalRange := last.High - last.Low

	// 1. Shooting Star / Pin Bar (Strong Reversal for Shorts)
	// Logic: Long upper wick (>2x body), small body
	if upperWick > body*2 && upperWick > totalRange*0.5 {
		return true, "Shooting Star (Eagle Eye)", 4
	}

	// 2. Bearish Engulfing Logic (Simplified)
	prev := candles[len(candles)-2]
	isGreen := prev.Close > prev.Open
	isRed := last.Close < last.Open

	if isGreen && isRed {
		if last.Open > prev.Close && last.Close < prev.Open {
			return true, "Bearish Engulfing", 5
		}
	}

	return false, "", 0
}
