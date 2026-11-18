package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"math"
	"time"
)

// DayConcentration prevents a hundred overlapping 5-minute signals from
// being passed off as one hundred independent market regimes.
type DayConcentration struct {
	Window                   string   `json:"window"`
	ResolvedObservations     int      `json:"resolved_observations"`
	IndependentUTCDays       int      `json:"independent_utc_days"`
	LargestSingleDaySamples  int      `json:"largest_single_day_samples"`
	LargestSingleDaySharePct *float64 `json:"largest_single_day_share_pct,omitempty"`
	RequiredDays             int      `json:"required_days"`
	Status                   string   `json:"status"`
}
type RegimeReport struct {
	Windows           []DayConcentration `json:"windows"`
	TradingAuthorized bool               `json:"trading_authorized"`
	Limitation        string             `json:"limitation"`
}

func RegimeAudit(samples []experiment.Sample, now time.Time) RegimeReport {
	out := RegimeReport{TradingAuthorized: false, Limitation: "UTC day count is a conservative clustering diagnostic, NOT proof that observations from different days are independent. A positive mean on one day is not a validated edge."}
	out.Windows = []DayConcentration{{Window: "5m", RequiredDays: 10, Status: "INSUFFICIENT_DISTINCT_DAYS"}, {Window: "15m", RequiredDays: 10, Status: "INSUFFICIENT_DISTINCT_DAYS"}}
	counts := [2]map[string]int{{}, {}}
	seen := map[string]bool{}
	for _, s := range samples {
		if s.Slug == "" || seen[s.Slug] || AuditIntegrity([]experiment.Sample{s}, now).Valid != 1 {
			continue
		}
		seen[s.Slug] = true
		if s.Winner == "" || s.ResolvedAt.IsZero() {
			continue
		}
		w := 0
		if s.WindowSeconds == 900 {
			w = 1
		}
		counts[w][s.CapturedAt.UTC().Format("2006-01-02")]++
	}
	for k := range out.Windows {
		r := &out.Windows[k]
		r.IndependentUTCDays = len(counts[k])
		for _, n := range counts[k] {
			r.ResolvedObservations += n
			if n > r.LargestSingleDaySamples {
				r.LargestSingleDaySamples = n
			}
		}
		if r.ResolvedObservations > 0 {
			p := math.Round(10000*float64(r.LargestSingleDaySamples)/float64(r.ResolvedObservations)) / 100
			r.LargestSingleDaySharePct = &p
		}
		if r.IndependentUTCDays >= r.RequiredDays && r.ResolvedObservations >= 100 {
			r.Status = "RESEARCH_DAY_DIVERSITY_PRESENT_NOT_INDEPENDENT_PROOF"
		}
	}
	return out
}
