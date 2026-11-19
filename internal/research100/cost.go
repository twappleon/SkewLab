package research100

import (
	"github.com/twappleon/SkewLab/internal/fees"
	"math"
	"time"
)

// EntryCost is for one hypothetical share held to official resolution.
// Platform fees plus an extra user-defined reserve are NOT quoted fills.
type EntryCost struct {
	Mode                        string  `json:"mode"`
	Ask                         float64 `json:"ask"`
	ModeledCostUSD              float64 `json:"modeled_cost_usd"`
	BreakEvenWinRatePct         float64 `json:"break_even_win_rate_pct"`
	MinimumWinRateWithBufferPct float64 `json:"minimum_win_rate_with_buffer_pct"`
	Status                      string  `json:"status"`
}

func CostToResolution(ask, slippage, extraBps, bufferPct float64, at time.Time) EntryCost {
	c := EntryCost{Mode: "HYPOTHETICAL_HOLD_TO_RESOLUTION", Ask: ask, Status: "INVALID_COST_INPUT"}
	if !(ask > 0 && ask < 1) || !finite(ask) || slippage < 0 || extraBps < 0 || bufferPct < 0 || !finite(slippage) || !finite(extraBps) || !finite(bufferPct) || at.IsZero() {
		return c
	}
	p := ask + slippage
	if p >= 1 {
		return c
	}
	cost := fees.BufferedCost(1, p, extraBps, at)
	if !finite(cost) || cost <= 0 {
		return c
	}
	c.ModeledCostUSD = math.Round(cost*1000000) / 1000000
	c.BreakEvenWinRatePct = 100 * cost
	c.MinimumWinRateWithBufferPct = 100 * cost * (1 + bufferPct/100)
	c.Status = "COST_MODEL_ONLY_NOT_EXECUTABLE"
	if cost >= 1 || c.MinimumWinRateWithBufferPct >= 100 {
		c.Status = "NO_FEASIBLE_BINARY_EDGE"
	}
	return c
}
