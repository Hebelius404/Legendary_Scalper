package exchange

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
)

// BinanceClient implements ExchangeClient for Binance Futures
type BinanceClient struct {
	client *futures.Client
}

// NewBinanceClient creates a new client instance
func NewBinanceClient(apiKey, apiSecret string, useTestnet bool) *BinanceClient {
	futures.UseTestnet = useTestnet
	client := binance.NewFuturesClient(apiKey, apiSecret)
	return &BinanceClient{client: client}
}

// GetServerTime returns exchange server time
func (b *BinanceClient) GetServerTime(ctx context.Context) (int64, error) {
	return b.client.NewServerTimeService().Do(ctx)
}

// GetUSDTBalance returns the available USDT balance in Futures wallet
func (b *BinanceClient) GetUSDTBalance(ctx context.Context) (float64, error) {
	acc, err := b.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return 0, err
	}

	for _, asset := range acc.Assets {
		if asset.Asset == "USDT" {
			// Convert string to float
			balance, err := strconv.ParseFloat(asset.WalletBalance, 64)
			if err != nil {
				return 0, fmt.Errorf("error parsing balance: %v", err)
			}
			return balance, nil
		}
	}

	return 0, fmt.Errorf("USDT asset not found in account")
}

// GetPrice returns the current price for a symbol
func (b *BinanceClient) GetPrice(ctx context.Context, symbol string) (float64, error) {
	prices, err := b.client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		return 0, err
	}

	if len(prices) == 0 {
		return 0, fmt.Errorf("no price data for symbol %s", symbol)
	}

	price, err := strconv.ParseFloat(prices[0].Price, 64)
	if err != nil {
		return 0, err
	}

	return price, nil
}

// NOTE: Get24hTicker and GetKlines are likely in market_data.go or another file if this project was split.
// The build error said they were "already declared".
// So I will OMIT them here to fix the build conflict.

// PlaceEntryOrder executes a MARKET entry
func (b *BinanceClient) PlaceEntryOrder(ctx context.Context, symbol, side string, quantity float64) (float64, error) {
	// 1. Get Step Size for Precision
	stepSize, err := b.getStepSize(ctx, symbol)
	if err != nil {
		log.Printf("⚠️ Failed to get step size for %s, defaulting to 1.0: %v", symbol, err)
		stepSize = 1.0
	}

	// 2. Adjust Quantity (Round down to step size)
	if stepSize > 0 {
		quantity = math.Floor(quantity/stepSize) * stepSize
	}

	// Determine precision for formatting string
	precision := 0
	if stepSize < 1 && stepSize > 0 {
		precision = int(math.Round(-math.Log10(stepSize)))
	}
	qtyStr := fmt.Sprintf("%.*f", precision, quantity)

	log.Printf("🛠️ Adjusting Quantity: -> %s (Step: %.6f)", qtyStr, stepSize)

	// Convert side
	var orderSide futures.SideType
	if side == "BUY" {
		orderSide = futures.SideTypeBuy
	} else {
		orderSide = futures.SideTypeSell
	}

	order, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(orderSide).
		Type(futures.OrderTypeMarket).
		Quantity(qtyStr).
		Do(ctx)
	if err != nil {
		return 0, err
	}

	if order.AvgPrice == "" {
		return 0, nil
	}
	avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)
	return avgPrice, nil
}

// getStepSize fetches the LOT_SIZE filter for a symbol
func (b *BinanceClient) getStepSize(ctx context.Context, symbol string) (float64, error) {
	info, err := b.client.NewExchangeInfoService().Do(ctx)
	if err != nil {
		return 0, err
	}

	// Parse LOT_SIZE filter
	for _, s := range info.Symbols {
		if s.Symbol == symbol {
			for _, f := range s.Filters {
				if f["filterType"] == "LOT_SIZE" {
					stepSizeStr, ok := f["stepSize"].(string)
					if !ok {
						return 1.0, fmt.Errorf("stepSize not a string")
					}
					return strconv.ParseFloat(stepSizeStr, 64)
				}
			}
		}
	}
	return 1.0, nil // Default fallback
}

// PlaceStopLoss executes a STOP_MARKET order
func (b *BinanceClient) PlaceStopLoss(ctx context.Context, symbol, side string, stopPrice float64) error {
	var orderSide futures.SideType
	if side == "BUY" { // Closing a SHORT
		orderSide = futures.SideTypeBuy
	} else { // Closing a LONG
		orderSide = futures.SideTypeSell
	}

	stopPriceStr := fmt.Sprintf("%.4f", stopPrice)

	_, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(orderSide).
		Type(futures.OrderTypeStopMarket).
		StopPrice(stopPriceStr).
		ClosePosition(true). // Close everything
		Do(ctx)
	return err
}

// PlaceTakeProfit executes a TAKE_PROFIT_MARKET order
func (b *BinanceClient) PlaceTakeProfit(ctx context.Context, symbol, side string, stopPrice float64) error {
	var orderSide futures.SideType
	if side == "BUY" {
		orderSide = futures.SideTypeBuy
	} else {
		orderSide = futures.SideTypeSell
	}

	stopPriceStr := fmt.Sprintf("%.4f", stopPrice)

	_, err := b.client.NewCreateOrderService().
		Symbol(symbol).
		Side(orderSide).
		Type(futures.OrderTypeTakeProfitMarket).
		StopPrice(stopPriceStr).
		ClosePosition(true).
		Do(ctx)
	return err
}

// GetPosition returns the current position details for a symbol
func (b *BinanceClient) GetPosition(ctx context.Context, symbol string) (*futures.PositionRisk, error) {
	positions, err := b.client.NewGetPositionRiskService().Symbol(symbol).Do(ctx)
	if err != nil {
		return nil, err
	}
	if len(positions) == 0 {
		return nil, fmt.Errorf("no position found for %s", symbol)
	}
	// Return pointer to copy or properly handle
	// positions is []*PositionRisk or []PositionRisk depending on SDK version
	// go-binance/v2 futures usually returns []*PositionRisk
	return positions[0], nil
}

// GetAllOpenPositions returns all positions with non-zero size
func (b *BinanceClient) GetAllOpenPositions(ctx context.Context) ([]*futures.PositionRisk, error) {
	positions, err := b.client.NewGetPositionRiskService().Do(ctx)
	if err != nil {
		return nil, err
	}
	var active []*futures.PositionRisk
	for _, p := range positions {
		// p is *futures.PositionRisk (assuming SDK returns slice of pointers)
		// If it's a slice of structs, p is a copy of the struct.
		// Let's verify SDK: futures/position_risk_service.go -> Do() returns ([]*PositionRisk, error)
		// So `positions` is `[]*PositionRisk`.
		// `p` is `*PositionRisk`.

		amt, _ := strconv.ParseFloat(p.PositionAmt, 64)
		if math.Abs(amt) > 0 {
			active = append(active, p)
		}
	}
	return active, nil
}
