// Package research100 provides deterministic, offline audit gates for SkewLab.
// It does not fetch live data, execute trades, or mutate paper balances.
package research100

import (
	"errors"
	"math"
	"time"
)

// Observation records an outcome-known-after-expiry comparison. The quote and
// signals were captured before resolution. A resolved label never implies
// independently verified settlement oracle data.
type Observation struct {
	ID         string    `json:"id"`
	Market     string    `json:"market"`
	Side       string    `json:"side"`
	CapturedAt time.Time `json:"captured_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	ResolvedAt time.Time `json:"resolved_at,omitempty"`
	Winner     string    `json:"winner,omitempty"`
	Bid        float64   `json:"bid"`
	Ask        float64   `json:"ask"`
	BidDepth   float64   `json:"bid_depth"`
	AskDepth   float64   `json:"ask_depth"`
	QuoteAt    time.Time `json:"quote_at"`
	Momentum   float64   `json:"momentum"`
	Skew       float64   `json:"skew"`
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Validate fails closed on contaminated prices, chronology, or labels.
func (o Observation) Validate() error {
	if o.ID == "" || (o.Market != "5m" && o.Market != "15m") || (o.Side != "UP" && o.Side != "DOWN") {
		return errors.New("invalid market identity or side")
	}
	if o.CapturedAt.IsZero() || !o.CapturedAt.Before(o.ExpiresAt) || o.QuoteAt.IsZero() || o.QuoteAt.After(o.CapturedAt) || o.CapturedAt.Sub(o.QuoteAt) > 10*time.Second {
		return errors.New("invalid or stale quote chronology")
	}
	if !finite(o.Bid) || !finite(o.Ask) || !finite(o.BidDepth) || !finite(o.AskDepth) || !finite(o.Momentum) || !finite(o.Skew) || o.Bid <= 0 || o.Ask >= 1 || o.Ask <= o.Bid || o.BidDepth <= 0 || o.AskDepth <= 0 || math.Abs(o.Skew) > 1.000001 {
		return errors.New("invalid quote or signal")
	}
	if o.Winner == "" {
		if !o.ResolvedAt.IsZero() {
			return errors.New("unlabeled observation with resolution time")
		}
		return nil
	}
	if o.Winner != "UP" && o.Winner != "DOWN" {
		return errors.New("invalid outcome")
	}
	if o.ResolvedAt.IsZero() || o.ResolvedAt.Before(o.ExpiresAt) {
		return errors.New("outcome known before expiry")
	}
	return nil
}
