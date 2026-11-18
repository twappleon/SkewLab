package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"testing"
	"time"
)

func TestDisagreementRequiresPastPublishedOutcomes(t *testing.T) {
	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	data := make([]experiment.Sample, 0, 103)
	for i := 0; i < 100; i++ {
		s := fixture(base.Add(time.Duration(i*6)*time.Minute), 300, "DOWN")
		data = append(data, s)
	}
	target := fixture(base.Add(100*6*time.Minute), 300, "UP")
	data = append(data, target)
	r := DisagreementAudit(data, base.Add(110*6*time.Minute))
	w := r.Windows[0]
	if w.ProspectiveDisagreements != 1 || w.SettledProspective != 1 || w.MeanIncrementalUSD == nil || *w.MeanIncrementalUSD <= 0 || w.IndependentUTCDays != 1 || r.TradingAuthorized {
		t.Fatalf("bad paired comparison: %+v", w)
	}
	for i := range data {
		if i < 100 {
			data[i].ResolvedAt = base.Add(2000 * time.Minute)
		}
	}
	r = DisagreementAudit(data, base.Add(2100*time.Minute))
	if r.Windows[0].ProspectiveDisagreements != 0 {
		t.Fatalf("future label leak: %+v", r.Windows[0])
	}
}
func TestDisagreementNoInventedSamples(t *testing.T) {
	r := DisagreementAudit(nil, time.Now())
	if r.Windows[0].MeanIncrementalUSD != nil || r.Windows[0].SettledProspective != 0 || r.TradingAuthorized {
		t.Fatal(r)
	}
}
