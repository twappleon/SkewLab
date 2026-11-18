package polymarket

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Spot struct {
	Price  float64 `json:"price"`
	Source string  `json:"source"`
}

// validSpot refuses malformed or nonsensical prices rather than manufacturing
// a BTC price from a failed provider. This is NOT a Chainlink settlement rate.
func validSpot(price string) (float64, error) {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil || p <= 100 || p > 10000000 || math.IsNaN(p) || math.IsInf(p, 0) {
		return 0, errors.New("bad/missing positive BTC price")
	}
	return p, nil
}

func (c *Client) fetchSpot(ctx context.Context, provider int) (Spot, error) {
	switch provider {
	case 0:
		var v struct {
			Price string `json:"price"`
		}
		if err := c.get(ctx, "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT", &v); err != nil {
			return Spot{}, err
		}
		p, e := validSpot(v.Price)
		return Spot{p, "Binance BTCUSDT spot"}, e
	case 1:
		var v struct {
			Price string `json:"price"`
		}
		if err := c.get(ctx, "https://api.exchange.coinbase.com/products/BTC-USD/ticker", &v); err != nil {
			return Spot{}, err
		}
		p, e := validSpot(v.Price)
		return Spot{p, "Coinbase BTC-USD spot"}, e
	case 2:
		var v struct {
			Error  []string `json:"error"`
			Result map[string]struct {
				Close []string `json:"c"`
			} `json:"result"`
		}
		if err := c.get(ctx, "https://api.kraken.com/0/public/Ticker?pair=XBTUSD", &v); err != nil {
			return Spot{}, err
		}
		if len(v.Error) > 0 {
			return Spot{}, fmt.Errorf("Kraken API error: %s", strings.Join(v.Error, ","))
		}
		for _, entry := range v.Result {
			if len(entry.Close) > 0 {
				p, err := validSpot(entry.Close[0])
				if err == nil {
					return Spot{p, "Kraken XBTUSD spot"}, nil
				}
			}
		}
		return Spot{}, errors.New("Kraken missing valid last trade")
	case 3:
		var v struct {
			Last      string `json:"last"`
			Timestamp string `json:"timestamp"`
		}
		if err := c.get(ctx, "https://www.bitstamp.net/api/v2/ticker/btcusd/", &v); err != nil {
			return Spot{}, err
		}
		p, err := validSpot(v.Last)
		if err != nil {
			return Spot{}, err
		}
		seconds, err := strconv.ParseInt(v.Timestamp, 10, 64)
		if err != nil || seconds <= 0 || time.Since(time.Unix(seconds, 0)) > 90*time.Second || time.Until(time.Unix(seconds, 0)) > 5*time.Second {
			return Spot{}, errors.New("Bitstamp ticker timestamp missing/stale/future")
		}
		return Spot{p, "Bitstamp BTCUSD spot"}, nil
	default:
		return Spot{}, errors.New("unsupported spot provider")
	}
}

// BTC sticks with the first working venue across calls. Switching to a
// fallback is permitted but the strategy must clear its momentum history on
// Source change; combining venues within one 30/60s window is invalid.
func (c *Client) BTC(ctx context.Context) (Spot, error) {
	c.spotMu.Lock()
	defer c.spotMu.Unlock()
	first := c.spotPreferred
	var order []int
	order = append(order, first)
	for i := 0; i < 4; i++ {
		if i != first {
			order = append(order, i)
		}
	}
	var failures []string
	names := []string{"Binance", "Coinbase", "Kraken", "Bitstamp"}
	for _, index := range order {
		spot, err := c.fetchSpot(ctx, index)
		if err == nil {
			c.spotPreferred = index
			return spot, nil
		}
		failures = append(failures, fmt.Sprintf("%s=%v", names[index], err))
		if ctx.Err() != nil {
			break
		}
	}
	return Spot{}, fmt.Errorf("BTC public spot feeds failed (%s)", strings.Join(failures, "; "))
}

// This client has no wallet, order signer, POST /order or trading credentials.
