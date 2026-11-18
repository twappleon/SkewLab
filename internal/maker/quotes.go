// Package maker emits non-executable research quotes. Public BBO snapshots
// cannot identify the position of hypothetical limit orders in a real queue.
// Consequently quotes are NEVER assumed filled and cannot affect equity.
package maker

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
	"math"
	"time"
)

type Proposal struct {
	Market  string  `json:"market"`
	Side    string  `json:"side"`
	Bid     float64 `json:"bid"`
	Ask     float64 `json:"ask"`
	Spread  float64 `json:"potential_spread"`
	SizeUSD float64 `json:"maximum_quote_notional_usd"`
	Status  string  `json:"status"`
	Reason  string  `json:"reason"`
}

type Config struct {
	Tick              float64
	MinObservedSpread float64
	QuotedNotionalUSD float64
	MaxSecondsToEnd   int
	MinSecondsToEnd   int
}

func DefaultConfig() Config {
	return Config{Tick: .001, MinObservedSpread: .014, QuotedNotionalUSD: 10, MinSecondsToEnd: 70}
}

// Plan emits bid/ask research suggestions only for the directional token
// supported by a valid momentum and semivariance signal. It never places or
// simulates a maker fill without a trade stream and queue model.
func Plan(m polymarket.Market, signal strategy.Result, now time.Time, c Config) Proposal {
	p := Proposal{Market: m.Slug, Status: "REJECTED", Reason: "no suitable market / signal"}
	if !m.Active || !signal.Ready || (signal.Direction != "UP" && signal.Direction != "DOWN") {
		return p
	}
	p.Side = signal.Direction
	if now.Add(time.Duration(c.MinSecondsToEnd) * time.Second).After(m.End) {
		p.Reason = "near expiry"
		return p
	}
	q := m.Up
	if p.Side == "DOWN" {
		q = m.Down
	}
	if q.Bid <= 0 || q.Ask <= 0 || q.Bid >= q.Ask || q.Ask-q.Bid < c.MinObservedSpread ||
		math.IsNaN(q.Bid) || math.IsNaN(q.Ask) || q.BidDepth <= 0 || q.AskDepth <= 0 {
		p.Reason = "invalid or tight public order book"
		return p
	}
	// Actual CLOB tick size overrides research fallback. Missing tick size
	// means no executable quote can be recommended.
	tick := q.TickSize
	if tick <= 0 || c.QuotedNotionalUSD <= 0 {
		p.Reason = "invalid maker configuration"
		return p
	}
	p.Bid = math.Round((q.Bid+tick)/tick) * tick
	p.Ask = math.Round((q.Ask-tick)/tick) * tick
	p.Bid = math.Round(p.Bid*1e6) / 1e6
	p.Ask = math.Round(p.Ask*1e6) / 1e6
	if p.Ask <= p.Bid || p.Ask >= q.Ask || p.Bid <= q.Bid || p.Bid >= q.Ask || p.Ask <= q.Bid {
		p.Status = "REJECTED"
		p.Reason = "price would cross or not improve the book"
		return p
	}
	p.Spread = p.Ask - p.Bid
	p.SizeUSD = c.QuotedNotionalUSD
	p.Status = "INDICATIVE_ONLY_NO_FILLS"
	p.Reason = "NOT an order or fill; sell quote requires inventory, queue priority and adverse selection unknown"
	return p
}
