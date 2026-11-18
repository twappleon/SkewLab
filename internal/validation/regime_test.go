package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"testing"
	"time"
)

func TestSingleDayConcentrationDoesNotClaimIndependence(t *testing.T) {
	day := time.Date(2026, 11, 2, 1, 0, 0, 0, time.UTC)
	var xs []experiment.Sample
	for i := 0; i < 100; i++ {
		at := day.Add(time.Duration(i*6) * time.Minute)
		s := fixture(at, 300, "UP")
		xs = append(xs, s)
	}
	r := RegimeAudit(xs, day.Add(30*time.Hour))
	w := r.Windows[0]
	if w.ResolvedObservations != 100 || w.IndependentUTCDays != 1 || w.LargestSingleDaySharePct == nil || *w.LargestSingleDaySharePct != 100 || w.Status != "INSUFFICIENT_DISTINCT_DAYS" || r.TradingAuthorized {
		t.Fatalf("misleading sample count: %+v", w)
	}
}
func TestDayReportNoPretendEvidence(t *testing.T) {
	r := RegimeAudit(nil, time.Now())
	if r.Windows[0].LargestSingleDaySharePct != nil || r.Windows[0].IndependentUTCDays != 0 {
		t.Fatal(r)
	}
}
