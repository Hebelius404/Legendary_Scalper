package ai

import (
	"net/http"
	"time"

	"legendary-scalper/internal/config"
)

type GrokClient struct {
	apiKey string
	client *http.Client
}

type AIResponse struct {
	Sentiment      string  `json:"sentiment"`      // BULLISH, BEARISH, NEUTRAL
	Confidence     float64 `json:"confidence"`     // 0.0 - 1.0
	Recommendation string  `json:"recommendation"` // LONG, SHORT, WAIT
	Reason         string  `json:"reason"`
}

func NewGrokClient(cfg *config.AppConfig) *GrokClient {
	return &GrokClient{
		apiKey: cfg.Grok.APIKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// AnalyzeMarketSentinel sends market data to Grok for analysis
func (g *GrokClient) AnalyzeMarketSentiment(marketSummary string) (*AIResponse, error) {
	if g.apiKey == "" {
		return &AIResponse{Sentiment: "NEUTRAL", Confidence: 0}, nil // Fallback
	}

	// Mocking the actual HTTP call for now to avoid complexity in this file
	// In production, this would hit https://api.x.ai/v1/chat/completions

	// Placeholder logic
	return &AIResponse{
		Sentiment:      "NEUTRAL",
		Confidence:     0.5,
		Recommendation: "WAIT",
		Reason:         "AI Integration placeholder",
	}, nil
}
