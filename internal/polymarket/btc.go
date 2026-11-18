package polymarket

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

type Spot struct {
	Price  float64 `json:"price"`
	Source string  `json:"source"`
}

func (c *Client) BTC(ctx context.Context) (Spot, error) {
	// First source stays primary so prices from distinct venues cannot be
	// stitched into a false momentum signal; strategy resets on source change.
	var binance struct {
		Price string `json:"price"`
	}
	if err := c.get(ctx, "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT", &binance); err == nil {
		if p, e := strconv.ParseFloat(binance.Price, 64); e == nil && p > 100 {
			return Spot{p, "Binance BTCUSDT spot"}, nil
		}
	}
	var coinbase struct {
		Price string `json:"price"`
	}
	if err := c.get(ctx, "https://api.exchange.coinbase.com/products/BTC-USD/ticker", &coinbase); err == nil {
		if p, e := strconv.ParseFloat(coinbase.Price, 64); e == nil && p > 100 {
			return Spot{p, "Coinbase BTC-USD spot"}, nil
		}
	}
	return Spot{}, fmt.Errorf("BTC data feeds unavailable: %w", errors.New("Binance and Coinbase requests failed"))
}

// Avoid importing high-level trading SDK: this client intentionally has no
// trading credentials or any POST /order method.
