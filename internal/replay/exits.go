package replay

import (
	"github.com/twappleon/SkewLab/internal/fees"
	"github.com/twappleon/SkewLab/internal/paper"
	"math"
	"time"
)

// ExitPolicy is a predeclared observational what-if. It cannot execute orders.
type ExitPolicy struct {
	Name          string  `json:"name"`
	TakeProfitPct float64 `json:"take_profit_pct"`
	StopLossPct   float64 `json:"stop_loss_pct"`
}

var OriginalPolicy = ExitPolicy{Name: "legacy_1pct_take_5pct_stop", TakeProfitPct: 1, StopLossPct: 5}
var AlternativePolicy = ExitPolicy{Name: "research_1pct_take_10pct_stop", TakeProfitPct: 1, StopLossPct: 10}

type ExitResult struct {
	Policy          string     `json:"policy"`
	Status          string     `json:"status"`
	Reason          string     `json:"reason,omitempty"`
	ObservedAt      *time.Time `json:"observed_at,omitempty"`
	ExitPrice       *float64   `json:"modeled_exit_price,omitempty"`
	ModeledPnLUSD   *float64   `json:"modeled_pnl_usd,omitempty"`
	CheckedQuotes   int        `json:"checked_quotes"`
	ValidExitQuotes int        `json:"valid_exit_quotes"`
}

// ReplayExit scans only snapshots captured AFTER the entry, in time order.
// A quote at a later market, a stale book, or inadequate bid depth never
// creates a simulated exit. The result is NOT an executable/real fill.
func ReplayExit(p paper.Position, frames []Frame, cfg paper.Config, policy ExitPolicy) ExitResult {
	out := ExitResult{Policy: policy.Name, Status: "NO_VERIFIABLE_EXIT"}
	if p.Market == "" || p.EntryTime.IsZero() || p.MarketEnd.IsZero() || !p.EntryTime.Before(p.MarketEnd) || !(p.EntryCost > 0) || !(p.Qty > 0) || !finite(p.EntryCost) || !finite(p.Qty) || cfg.SlippagePrice < 0 || cfg.FeeBps < 0 || !finite(cfg.FeeBps) || !finite(cfg.SlippagePrice) || policy.TakeProfitPct < 1 || policy.StopLossPct <= 0 || policy.Name == "" {
		out.Status = "INVALID_INPUT"
		return out
	}
	for _, frame := range frames {
		if !frame.Time.After(p.EntryTime) || !frame.Time.Before(p.MarketEnd) {
			continue
		}
		m, ok := marketFor(frame, p)
		if !ok || !m.Active {
			continue
		}
		out.CheckedQuotes++
		q := quoteFor(m, p.Side)
		if !validQuote(q, frame.Time) || q.BidDepth+1e-6 < p.Qty {
			continue
		}
		price := math.Max(0.001, q.Bid-cfg.SlippagePrice)
		if !(price > 0 && price < 1) {
			continue
		}
		net := fees.BufferedProceeds(p.Qty, price, cfg.FeeBps, frame.Time) - p.EntryCost
		if !finite(net) {
			continue
		}
		out.ValidExitQuotes++
		pnlPct := 100 * net / p.EntryCost
		nearEnd := frame.Time.Add(32 * time.Second).After(p.MarketEnd)
		// Mirror original 4s anti-churn delay, except near market expiry.
		if frame.Time.Sub(p.EntryTime) < 4*time.Second && !nearEnd {
			continue
		}
		reason := ""
		if nearEnd {
			reason = "pre_settlement"
		} else if pnlPct >= policy.TakeProfitPct {
			reason = "take_profit"
		} else if pnlPct <= -policy.StopLossPct {
			reason = "stop_loss"
		}
		if reason == "" {
			continue
		}
		t := frame.Time
		out.ObservedAt = &t
		out.ExitPrice = &price
		out.ModeledPnLUSD = &net
		out.Reason = reason
		out.Status = "MODELED_EXIT_FROM_RECORDED_QUOTE"
		return out
	}
	return out
}
