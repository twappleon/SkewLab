package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"math"
	"sort"
	"time"
)

type SettlementHealth struct {
	ValidObservations int      `json:"valid_observations"`
	Resolved          int      `json:"resolved"`
	Pending           int      `json:"pending"`
	Overdue           int      `json:"overdue"`
	ResolutionRatePct *float64 `json:"resolution_rate_pct,omitempty"`
	LagP50Seconds     *float64 `json:"lag_p50_seconds,omitempty"`
	LagP95Seconds     *float64 `json:"lag_p95_seconds,omitempty"`
	Status            string   `json:"status"`
}

// SettlementAudit uses only individually valid, previously recorded market
// observations. Missing outcomes are never inferred from spot prices.
func SettlementAudit(xs []experiment.Sample, now time.Time) SettlementHealth {
	a := SettlementHealth{Status: "NO_OBSERVATIONS"}
	seen := map[string]bool{}
	lags := []float64{}
	for _, s := range xs {
		if s.Slug == "" || seen[s.Slug] || s.CapturedAt.After(now) || s.MarketEnd.IsZero() || s.MarketEnd.Before(s.CapturedAt) || s.Policy != experiment.SamplingPolicy || (s.WindowSeconds != 300 && s.WindowSeconds != 900) {
			continue
		}
		seen[s.Slug] = true
		a.ValidObservations++
		if s.Winner == "" {
			a.Pending++
			if now.After(s.MarketEnd.Add(30 * time.Minute)) {
				a.Overdue++
			}
			continue
		}
		if (s.Winner != "UP" && s.Winner != "DOWN") || s.ResolvedAt.IsZero() || s.ResolvedAt.Before(s.MarketEnd) || s.ResolvedAt.After(now) {
			a.ValidObservations--
			continue
		}
		a.Resolved++
		lags = append(lags, s.ResolvedAt.Sub(s.MarketEnd).Seconds())
	}
	if a.ValidObservations > 0 {
		v := math.Round(10000*float64(a.Resolved)/float64(a.ValidObservations)) / 100
		a.ResolutionRatePct = &v
		a.Status = "COLLECTING"
		if a.Overdue > 0 {
			a.Status = "SETTLEMENT_BACKLOG"
		} else if a.Resolved >= 30 {
			a.Status = "OBSERVED_SETTLEMENTS_ONLY"
		}
	}
	if len(lags) > 0 {
		sort.Float64s(lags)
		p50 := lags[int(math.Ceil(.5*float64(len(lags))))-1]
		p95 := lags[int(math.Ceil(.95*float64(len(lags))))-1]
		a.LagP50Seconds = &p50
		a.LagP95Seconds = &p95
	}
	return a
}
