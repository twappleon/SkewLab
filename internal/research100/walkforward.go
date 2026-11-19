package research100

import (
	"sort"
	"time"
)

type WalkForward struct {
	ProspectiveCandidates      int     `json:"prospective_candidates"`
	EligibleResearchCandidates int     `json:"eligible_research_candidates"`
	ScoredResolved             int     `json:"scored_resolved"`
	Wins                       int     `json:"wins"`
	HypotheticalNetUSD         float64 `json:"hypothetical_net_per_share_sum_usd"`
	Status                     string  `json:"status"`
	TradingAuthorized          bool    `json:"trading_authorized"`
}

// RunWalkForward advances the known-outcome frontier by ResolvedAt and never
// uses a winning label before it was published. One share held to expiry is a
// research assumption, NOT the existing intraday stop-loss strategy.
func RunWalkForward(d Dataset, asOf time.Time) WalkForward {
	out := WalkForward{Status: "INSUFFICIENT_FORWARD_EVIDENCE", TradingAuthorized: false}
	resolved := make([]Observation, 0, len(d.Items))
	for _, o := range d.Items {
		if o.Winner != "" && !o.ResolvedAt.After(asOf) {
			resolved = append(resolved, o)
		}
	}
	sort.Slice(resolved, func(i, j int) bool { return resolved[i].ResolvedAt.Before(resolved[j].ResolvedAt) })
	type hist struct {
		samples, wins int
		days          map[string]bool
	}
	history := map[string]*hist{}
	idx := 0
	for _, o := range d.Items {
		if o.CapturedAt.After(asOf) {
			continue
		}
		for idx < len(resolved) && resolved[idx].ResolvedAt.Before(o.CapturedAt) {
			r := resolved[idx]
			key := r.Market + ":" + r.Side
			h := history[key]
			if h == nil {
				h = &hist{days: map[string]bool{}}
				history[key] = h
			}
			h.samples++
			if r.Winner == r.Side {
				h.wins++
			}
			h.days[r.CapturedAt.UTC().Format("2006-01-02")] = true
			idx++
		}
		out.ProspectiveCandidates++
		h := history[o.Market+":"+o.Side]
		if h == nil || h.samples < 100 || len(h.days) < 10 {
			continue
		}
		cost := CostToResolution(o.Ask, .002, 25, 2, o.CapturedAt)
		if cost.Status != "COST_MODEL_ONLY_NOT_EXECUTABLE" {
			continue
		}
		lower := wilsonLower95(h.wins, h.samples)
		if lower/cost.ModeledCostUSD-1 < .02 {
			continue
		}
		out.EligibleResearchCandidates++
		if o.Winner == "" || o.ResolvedAt.After(asOf) {
			continue
		}
		out.ScoredResolved++
		if o.Winner == o.Side {
			out.Wins++
			out.HypotheticalNetUSD += 1 - cost.ModeledCostUSD
		} else {
			out.HypotheticalNetUSD -= cost.ModeledCostUSD
		}
	}
	if out.ScoredResolved >= 30 {
		out.Status = "FORWARD_RESEARCH_SAMPLE_NOT_TRADE_APPROVAL"
	}
	return out
}
