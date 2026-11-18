package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"math"
	"time"
)

// Integrity checks the stored fixed-time observation archive. It never
// backfills missing markets or interprets observations as real fills.
type Integrity struct {
	Total           int            `json:"total"`
	Valid           int            `json:"valid"`
	Invalid         int            `json:"invalid"`
	Duplicates      int            `json:"duplicates"`
	FutureSamples   int            `json:"future_samples"`
	MissingOutcomes int            `json:"missing_outcomes"`
	ByWindow        map[string]int `json:"by_window"`
	Status          string         `json:"status"`
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func AuditIntegrity(xs []experiment.Sample, now time.Time) Integrity {
	a := Integrity{ByWindow: map[string]int{"5m": 0, "15m": 0}, Status: "NO_OBSERVATIONS"}
	seen := make(map[string]bool, len(xs))
	a.Total = len(xs)
	for _, s := range xs {
		if s.Slug == "" || seen[s.Slug] {
			a.Duplicates++
			a.Invalid++
			continue
		}
		seen[s.Slug] = true
		start := s.MarketEnd.Add(-time.Duration(s.WindowSeconds) * time.Second)
		elapsed := s.CapturedAt.Sub(start)
		if s.CapturedAt.After(now) || s.ResolvedAt.After(now) {
			a.FutureSamples++
			a.Invalid++
			continue
		}
		if (s.WindowSeconds != 300 && s.WindowSeconds != 900) || s.Policy != experiment.SamplingPolicy || s.ConditionID == "" || s.CapturedAt.IsZero() || s.MarketEnd.IsZero() || elapsed < 90*time.Second || elapsed > 120*time.Second || !s.CapturedAt.Before(s.MarketEnd) || s.UpQuoteAt.IsZero() || s.DownQuoteAt.IsZero() || s.CapturedAt.Sub(s.UpQuoteAt) > 10*time.Second || s.CapturedAt.Sub(s.DownQuoteAt) > 10*time.Second || s.UpQuoteAt.After(s.CapturedAt.Add(3*time.Second)) || s.DownQuoteAt.After(s.CapturedAt.Add(3*time.Second)) || !finite(s.UpAsk) || !finite(s.DownAsk) || !finite(s.UpBid) || !finite(s.DownBid) || !finite(s.MomentumPct) || !finite(s.Skew) || s.UpBid <= 0 || s.DownBid <= 0 || s.UpAsk <= s.UpBid || s.DownAsk <= s.DownBid || s.UpAsk >= 1 || s.DownAsk >= 1 || (s.Winner != "" && s.Winner != "UP" && s.Winner != "DOWN") || (s.Winner != "" && s.ResolvedAt.Before(s.MarketEnd)) || (s.Winner == "" && !s.ResolvedAt.IsZero()) {
			a.Invalid++
			continue
		}
		a.Valid++
		if s.WindowSeconds == 300 {
			a.ByWindow["5m"]++
		} else {
			a.ByWindow["15m"]++
		}
		if s.Winner == "" {
			a.MissingOutcomes++
		}
	}
	if a.Total > 0 {
		if a.Invalid > 0 {
			a.Status = "ARCHIVE_REVIEW_REQUIRED"
		} else {
			a.Status = "VALID_OBSERVED_RECORDS_ONLY"
		}
	}
	return a
}
