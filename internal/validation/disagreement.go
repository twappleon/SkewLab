package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"math"
	"sort"
	"time"
)

// Independent edge matters only when a rule disagrees with contemporaneous
// market prices. A matching rule cannot claim incremental directional alpha.
// Every compared outcome must be labeled AFTER its capture, and the 100-label
// history must be published strictly BEFORE the candidate's capture time.
type DisagreementWindow struct {
	Window                   string   `json:"window"`
	Recorded                 int      `json:"recorded"`
	MatchedSignalAndBaseline int      `json:"matched_signal_and_baseline"`
	DifferentDirection       int      `json:"different_direction"`
	HistoricalReady          int      `json:"historical_ready"`
	ProspectiveDisagreements int      `json:"prospective_disagreements"`
	SettledProspective       int      `json:"settled_prospective"`
	IndependentUTCDays       int      `json:"independent_utc_days"`
	MeanIncrementalUSD       *float64 `json:"mean_incremental_usd,omitempty"`
	Status                   string   `json:"status"`
}
type DisagreementReport struct {
	Windows           []DisagreementWindow `json:"windows"`
	TradingAuthorized bool                 `json:"trading_authorized"`
	Limitations       string               `json:"limitations"`
}

func DisagreementAudit(samples []experiment.Sample, now time.Time) DisagreementReport {
	out := DisagreementReport{TradingAuthorized: false, Limitations: "Only same-timestamp, fee-adjusted hypothetical hold-to-resolution outcomes. No market baseline outperformance is established without prior data, forward labels and independent days."}
	out.Windows = []DisagreementWindow{{Window: "5m", Status: "NO_FORWARD_DISAGREEMENTS"}, {Window: "15m", Status: "NO_FORWARD_DISAGREEMENTS"}}
	seen := map[string]bool{}
	ordered := []experiment.Sample{}
	published := []experiment.Sample{}
	for _, s := range samples {
		if s.Slug == "" || seen[s.Slug] || AuditIntegrity([]experiment.Sample{s}, now).Valid != 1 {
			continue
		}
		seen[s.Slug] = true
		ordered = append(ordered, s)
		if s.Winner != "" && !s.ResolvedAt.IsZero() {
			published = append(published, s)
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CapturedAt.Before(ordered[j].CapturedAt) })
	sort.Slice(published, func(i, j int) bool { return published[i].ResolvedAt.Before(published[j].ResolvedAt) })
	counts := [2]int{}
	index := 0
	sums := [2]float64{}
	days := [2]map[string]bool{{}, {}}
	for _, s := range ordered {
		for index < len(published) && published[index].ResolvedAt.Before(s.CapturedAt) {
			if published[index].WindowSeconds == 300 {
				counts[0]++
			} else {
				counts[1]++
			}
			index++
		}
		k := 0
		if s.WindowSeconds == 900 {
			k = 1
		}
		r := &out.Windows[k]
		r.Recorded++
		rule := pick(s, 1)
		baseline := pick(s, 2)
		if rule == "" || baseline == "" {
			continue
		}
		if rule == baseline {
			r.MatchedSignalAndBaseline++
			continue
		}
		r.DifferentDirection++
		if counts[k] < 100 {
			continue
		}
		r.HistoricalReady++
		r.ProspectiveDisagreements++
		if s.Winner == "" {
			continue
		}
		a, b := net(s, rule, .002, 25), net(s, baseline, .002, 25)
		if !finite(a) || !finite(b) {
			continue
		}
		r.SettledProspective++
		sums[k] += a - b
		days[k][s.CapturedAt.UTC().Format("2006-01-02")] = true
	}
	for k := range out.Windows {
		r := &out.Windows[k]
		r.IndependentUTCDays = len(days[k])
		if r.SettledProspective > 0 {
			v := math.Round(sums[k]/float64(r.SettledProspective)*10000) / 10000
			r.MeanIncrementalUSD = &v
		}
		if r.SettledProspective < 30 || r.IndependentUTCDays < 10 {
			r.Status = "INSUFFICIENT_FORWARD_INDEPENDENT_EVIDENCE"
		} else {
			r.Status = "RESEARCH_REVIEW_ONLY_NO_TRADING_APPROVAL"
		}
	}
	return out
}
