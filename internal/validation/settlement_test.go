package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"testing"
	"time"
)

func TestSettlementDoesNotInventUnresolvedOutcomes(t *testing.T) {
	now := time.Date(2026, 11, 18, 16, 0, 0, 0, time.UTC)
	a := fixture(now.Add(-2*time.Hour), 300, "UP")
	b := fixture(now.Add(-time.Hour), 300, "")
	b.Slug = "b"
	b.ResolvedAt = time.Time{}
	c := fixture(now.Add(-10*time.Minute), 900, "")
	c.Slug = "c"
	c.ResolvedAt = time.Time{}
	r := SettlementAudit([]experiment.Sample{a, b, c}, now)
	if r.Resolved != 1 || r.Pending != 2 || r.Overdue != 1 || r.ResolutionRatePct == nil || *r.ResolutionRatePct != 33.33 || r.Status != "SETTLEMENT_BACKLOG" {
		t.Fatalf("bad settlement report: %+v", r)
	}
	future := a
	future.Slug = "future"
	future.ResolvedAt = now.Add(time.Hour)
	r = SettlementAudit([]experiment.Sample{future}, now)
	if r.Resolved != 0 || r.ValidObservations != 0 {
		t.Fatalf("future outcome accepted: %+v", r)
	}
	if x := SettlementAudit(nil, now); x.LagP95Seconds != nil || x.ResolutionRatePct != nil {
		t.Fatal(x)
	}
}
