package replay

import (
	"github.com/twappleon/SkewLab/internal/fees"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"math"
	"time"
)

// CostStress is a same-quote hypothetical round trip, never a fill or strategy.
// It assumes top-of-book depth is available in BOTH directions.
type CostStress struct {
	Status                   string   `json:"status"`
	NormalLossPct            *float64 `json:"normal_loss_pct,omitempty"`
	DoubledSlippageLossPct   *float64 `json:"double_slippage_loss_pct,omitempty"`
	QuadrupledReserveLossPct *float64 `json:"fourfold_reserve_loss_pct,omitempty"`
	StopAlreadyConsumed      bool     `json:"stop_consumed_by_normal_friction"`
}

func EvaluateCostStress(q polymarket.Quote, now time.Time, qty, slippage, reserveBps, stopPct float64) CostStress {
	out := CostStress{Status: "NO_VALID_QUOTE_OR_DEPTH"}
	if !validQuote(q, now) || !finite(qty) || qty <= 0 || q.AskDepth < qty || q.BidDepth < qty || !finite(slippage) || slippage < 0 || !finite(reserveBps) || reserveBps < 0 || !finite(stopPct) || stopPct <= 0 {
		return out
	}
	sim := func(slip, bps float64) *float64 {
		in := q.Ask + slip
		outPx := q.Bid - slip
		if in >= 1 || outPx <= 0 {
			return nil
		}
		cost := fees.BufferedCost(qty, in, bps, now)
		if cost <= 0 || !finite(cost) {
			return nil
		}
		proceeds := fees.BufferedProceeds(qty, outPx, bps, now)
		if !finite(proceeds) {
			return nil
		}
		val := 100 * (cost - proceeds) / cost
		if !finite(val) || math.Abs(val) > 1000 {
			return nil
		}
		return &val
	}
	out.NormalLossPct = sim(slippage, reserveBps)
	out.DoubledSlippageLossPct = sim(slippage*2, reserveBps)
	out.QuadrupledReserveLossPct = sim(slippage, reserveBps*4)
	if out.NormalLossPct == nil || out.DoubledSlippageLossPct == nil || out.QuadrupledReserveLossPct == nil {
		return CostStress{Status: "INVALID_COST_ASSUMPTIONS"}
	}
	out.Status = "RESEARCH_ONLY"
	out.StopAlreadyConsumed = *out.NormalLossPct >= stopPct
	return out
}
