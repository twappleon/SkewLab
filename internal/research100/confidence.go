package research100

import (
	"math"
	"time"
)

type EdgeReview struct {
	Status                  string   `json:"status"`
	Samples                 int      `json:"samples"`
	Winners                 int      `json:"winners"`
	UTCObservationDays      int      `json:"utc_observation_days"`
	WilsonLower95           *float64 `json:"wilson_lower_95,omitempty"`
	ConservativeExpectedROI *float64 `json:"conservative_expected_roi,omitempty"`
	TradingAuthorized       bool     `json:"trading_authorized"`
}

func wilsonLower95(wins, total int) float64 {
	if total <= 0 {
		return 0
	}
	n := float64(total)
	p := float64(wins) / n
	z := 1.96
	d := 1 + z*z/n
	return math.Max(0, (p+z*z/(2*n)-z*math.Sqrt(p*(1-p)/n+z*z/(4*n*n)))/d)
}

// EvaluateEdge only uses officially labeled outcomes resolved BEFORE asOf.
// Day diversity is a minimal correlation check, not independence proof.
func EvaluateEdge(d Dataset, window, side string, asOf time.Time, entryAsk, slippage, extraBps float64) EdgeReview {
	out := EdgeReview{Status: "INSUFFICIENT_PRIOR_EVIDENCE", TradingAuthorized: false}
	if (window != "5m" && window != "15m") || (side != "UP" && side != "DOWN") || asOf.IsZero() {
		out.Status = "INVALID_RESEARCH_REQUEST"
		return out
	}
	cost := CostToResolution(entryAsk, slippage, extraBps, 2, asOf)
	if cost.Status != "COST_MODEL_ONLY_NOT_EXECUTABLE" {
		out.Status = "NO_FEASIBLE_ENTRY_COST"
		return out
	}
	days := map[string]bool{}
	for _, o := range d.Items {
		if o.Market != window || o.Side != side || o.Winner == "" || !o.ResolvedAt.Before(asOf) || !o.CapturedAt.Before(asOf) {
			continue
		}
		out.Samples++
		days[o.CapturedAt.UTC().Format("2006-01-02")] = true
		if o.Winner == side {
			out.Winners++
		}
	}
	out.UTCObservationDays = len(days)
	if out.Samples < 100 || out.UTCObservationDays < 10 {
		return out
	}
	lower := wilsonLower95(out.Winners, out.Samples)
	roi := lower/cost.ModeledCostUSD - 1
	out.WilsonLower95 = &lower
	out.ConservativeExpectedROI = &roi
	out.Status = "NO_CONSERVATIVE_EDGE"
	if roi >= .02 {
		out.Status = "RESEARCH_CANDIDATE_ONLY"
	}
	return out
}
