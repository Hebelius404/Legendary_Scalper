package scanner

import (
	"context"
	"fmt"
	"sort"

	"legendary-scalper/internal/config"
	"legendary-scalper/internal/exchange"
)

// MarketOpportunity represents a coin that meets criteria
type MarketOpportunity struct {
	Symbol     string
	Price      float64
	Change24h  float64 // Percentage
	Volume24h  float64 // in USDT
	RSI        float64
	Volatility float64
	Score      float64
}

// ScannerEngine manages market scanning
type ScannerEngine struct {
	cfg    *config.ScanningConfig
	client exchange.ExchangeClient
}

// NewScannerEngine creates a new scanner
func NewScannerEngine(cfg *config.ScanningConfig, client exchange.ExchangeClient) *ScannerEngine {
	return &ScannerEngine{cfg: cfg, client: client}
}

// ScanForOpportunities finds coins matching pump/volatility criteria
func (s *ScannerEngine) ScanForOpportunities(ctx context.Context) ([]MarketOpportunity, error) {
	// 1. Fetch 24h stats for all symbols
	tickers, err := s.client.Get24hTicker(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tickers: %v", err)
	}

	var candidates []MarketOpportunity

	for _, t := range tickers {
		// Filter 1: Quote Asset (USDT only)
		// Minimal check: symbol ends with "USDT"
		if len(t.Symbol) < 5 || t.Symbol[len(t.Symbol)-4:] != "USDT" {
			continue
		}

		// Filter 2: Blacklist
		if s.isBlacklisted(t.Symbol) {
			continue
		}

		// Filter 3: Minimum Volume
		if t.Volume < s.cfg.MinVolume {
			continue
		}

		// Filter 4: Pump/Volatility (Change > 30% for PumpDetector)
		// Only consider positive pumps for Shorting (Martingale style)
		// Or high volatility if we are just looking for action.
		// For now, let's implement the Pump Detector logic.
		// Need to access Strategy Config ideally, but we have ScanningConfig here.
		// Assuming we want coins with SIGNIFICANT movement.

		// If using Volatility Ranking, we just collect them all and sort.
		// If just Pump Detection, we check % change.

		candidates = append(candidates, MarketOpportunity{
			Symbol:    t.Symbol,
			Price:     t.LastPrice,
			Change24h: t.PriceChange,
			Volume24h: t.Volume,
			// RSI/Volatility calculated later for top candidates only to save API calls
		})
	}

	// Sort by 24h Change (Descending) - Find the biggest pumps
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Change24h > candidates[j].Change24h
	})

	// Limit to Top X to avoid processing too many
	limit := s.cfg.TopPairs
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	return candidates, nil
}

// isBlacklisted checks config blacklist
func (s *ScannerEngine) isBlacklisted(symbol string) bool {
	for _, b := range s.cfg.Blacklist {
		if b == symbol {
			return true
		}
	}
	return false
}
