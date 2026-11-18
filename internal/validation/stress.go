package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"math"
	"time"
)

type StressScenario struct {
	Name             string       `json:"name"`
	SlippagePerShare float64      `json:"slippage_per_share"`
	AdditionalFeeBps float64      `json:"additional_fee_bps"`
	Paired           PairedReport `json:"paired_results"`
}
type StressReport struct {
	Scenarios            []StressScenario `json:"scenarios"`
	MinimumMatched       int              `json:"minimum_matched"`
	AllScenariosPositive bool             `json:"all_scenarios_positive"`
	Status               string           `json:"status"`
	TradingAuthorized    bool             `json:"trading_authorized"`
}

// StressCompare fixes the historical cohort and resolution chronology, then
// re-prices already-observed snapshots with stronger hypothetical cost reserves.
// It cannot model actual fills or early stop/take-profit execution.
func StressCompare(samples []experiment.Sample, now time.Time) StressReport {
	cases := []struct {
		name      string
		slip, bps float64
	}{
		{"baseline", .002, 25}, {"wide_spread_reserve", .01, 60}, {"severe_liquidity_reserve", .02, 100},
	}
	out := StressReport{Status: "INSUFFICIENT_MATCHED_EVIDENCE", TradingAuthorized: false}
	all := true
	min := math.MaxInt
	for _, c := range cases {
		p := PairedCheck(samples, now, c.slip, c.bps)
		out.Scenarios = append(out.Scenarios, StressScenario{Name: c.name, SlippagePerShare: c.slip, AdditionalFeeBps: c.bps, Paired: p})
		for _, window := range []string{"5m", "15m"} {
			for _, rule := range p.ByWindow[window] {
				if rule.MatchedSettled < min {
					min = rule.MatchedSettled
				}
				if rule.ConservativeDayBlockLowerUSD == nil || *rule.ConservativeDayBlockLowerUSD <= 0 {
					all = false
				}
			}
		}
	}
	if min == math.MaxInt {
		min = 0
	}
	out.MinimumMatched = min
	out.AllScenariosPositive = all
	if all {
		out.Status = "EXPLORATORY_STRESS_PASS_NOT_TRADING_APPROVAL"
	}
	return out
}
