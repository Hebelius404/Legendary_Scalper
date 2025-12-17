package exchange

import (
	"context"

	"github.com/adshao/go-binance/v2/futures"
)

type Ticker24h struct {
	Symbol      string
	PriceChange float64 // Percent
	LastPrice   float64
	Volume      float64 // Quote Volume (USDT)
}

type Kline struct {
	OpenTime int64
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   float64
}

// ExchangeClient defines the methods required for any exchange implementation
type ExchangeClient interface {
	// core
	GetServerTime(ctx context.Context) (int64, error)
	GetUSDTBalance(ctx context.Context) (float64, error)

	// market data
	GetPrice(ctx context.Context, symbol string) (float64, error)
	Get24hTicker(ctx context.Context) ([]Ticker24h, error)
	GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error)

	// trading
	PlaceEntryOrder(ctx context.Context, symbol, side string, quantity float64) (float64, error)
	PlaceStopLoss(ctx context.Context, symbol, side string, stopPrice float64) error
	PlaceTakeProfit(ctx context.Context, symbol, side string, stopPrice float64) error

	// position management
	GetPosition(ctx context.Context, symbol string) (*futures.PositionRisk, error)
	GetAllOpenPositions(ctx context.Context) ([]*futures.PositionRisk, error)
}
