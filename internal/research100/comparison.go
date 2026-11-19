package research100

import (
	"time"
)

type RuleStat struct {
	Name                   string  `json:"name"`
	EvaluatedSameSnapshots int     `json:"evaluated_same_snapshots"`
	Decisions              int     `json:"decisions"`
	Wins                   int     `json:"wins"`
	NetOneShareUSD         float64 `json:"net_one_share_usd"`
}
type PairedComparison struct {
	OutcomeLabeled int        `json:"outcome_labeled"`
	Rules          []RuleStat `json:"rules"`
	Status         string     `json:"status"`
}

// CompareRules makes no equal-trade-count assumption: abstentions stay visible.
// The always-buy-sampled-side rule is NOT a fair two-sided market baseline;
// only use this for detecting the incremental effect of signal filtering.
func CompareRules(d Dataset, asOf time.Time) PairedComparison {
	out := PairedComparison{Status: "SELECTED_SIDE_RESEARCH_ONLY", Rules: []RuleStat{{Name: "buy_sampled_side"}, {Name: "momentum_filter"}, {Name: "momentum_plus_skew_filter"}}}
	for _, o := range d.Items {
		if o.Winner == "" || o.ResolvedAt.After(asOf) {
			continue
		}
		c := CostToResolution(o.Ask, .002, 25, 0, o.CapturedAt)
		if c.Status != "COST_MODEL_ONLY_NOT_EXECUTABLE" {
			continue
		}
		out.OutcomeLabeled++
		signalAligned := (o.Side == "UP" && o.Momentum >= .02) || (o.Side == "DOWN" && o.Momentum <= -.02)
		skewAligned := (o.Side == "UP" && o.Skew >= .2) || (o.Side == "DOWN" && o.Skew <= -.2)
		for i := range out.Rules {
			r := &out.Rules[i]
			r.EvaluatedSameSnapshots++
			if i == 1 && !signalAligned || i == 2 && !(signalAligned && skewAligned) {
				continue
			}
			r.Decisions++
			payoff := 0.0
			if o.Winner == o.Side {
				r.Wins++
				payoff = 1
			}
			r.NetOneShareUSD += payoff - c.ModeledCostUSD
		}
	}
	if out.OutcomeLabeled == 0 {
		out.Status = "NO_RESOLVED_SELECTED_SIDE_SAMPLES"
	}
	return out
}
