package exchange

import (
	"context"
	"strconv"
)

// Get24hTicker returns 24h stats for all symbols
func (b *BinanceClient) Get24hTicker(ctx context.Context) ([]Ticker24h, error) {
	stats, err := b.client.NewListPriceChangeStatsService().Do(ctx)
	if err != nil {
		return nil, err
	}

	var result []Ticker24h
	for _, s := range stats {
		priceChange, _ := strconv.ParseFloat(s.PriceChangePercent, 64)
		lastPrice, _ := strconv.ParseFloat(s.LastPrice, 64)
		volume, _ := strconv.ParseFloat(s.QuoteVolume, 64)

		result = append(result, Ticker24h{
			Symbol:      s.Symbol,
			PriceChange: priceChange,
			LastPrice:   lastPrice,
			Volume:      volume,
		})
	}
	return result, nil
}

// GetKlines returns candlestick data
func (b *BinanceClient) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	klines, err := b.client.NewKlinesService().
		Symbol(symbol).
		Interval(interval).
		Limit(limit).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	var result []Kline
	for _, k := range klines {
		open, _ := strconv.ParseFloat(k.Open, 64)
		high, _ := strconv.ParseFloat(k.High, 64)
		low, _ := strconv.ParseFloat(k.Low, 64)
		closePrice, _ := strconv.ParseFloat(k.Close, 64)
		volume, _ := strconv.ParseFloat(k.Volume, 64)

		result = append(result, Kline{
			OpenTime: k.OpenTime,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closePrice,
			Volume:   volume,
		})
	}
	return result, nil
}
