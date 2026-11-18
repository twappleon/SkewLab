package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"testing"
	"time"
)

func fixture(at time.Time, seconds int, side string) experiment.Sample {
	start := at.Add(-100 * time.Second)
	end := start.Add(time.Duration(seconds) * time.Second)
	return experiment.Sample{Slug: at.Format(time.RFC3339Nano), ConditionID: "cid", Policy: experiment.SamplingPolicy, WindowSeconds: seconds, CapturedAt: at, MarketEnd: end, UpQuoteAt: at, DownQuoteAt: at, UpBid: .48, UpAsk: .50, UpAskDepth: 100, DownBid: .49, DownAsk: .51, DownAskDepth: 100, MomentumPct: .04, Skew: .3, BTCSpotSource: "test", Winner: side, ResolvedAt: end.Add(time.Minute)}
}
func TestIntegrityRejectsBadChronologyDuplicatesAndFuture(t *testing.T) {
	now := time.Date(2026, 11, 18, 14, 0, 0, 0, time.UTC)
	a := fixture(now.Add(-time.Hour), 300, "UP")
	b := a
	b.Slug = "another"
	b.Winner = ""
	b.ResolvedAt = time.Time{}
	c := a
	c.Slug = "third"
	c.CapturedAt = now.Add(time.Minute)
	r := AuditIntegrity([]experiment.Sample{a, b, a, c}, now)
	if r.Total != 4 || r.Valid != 2 || r.Invalid != 2 || r.Duplicates != 1 || r.FutureSamples != 1 || r.MissingOutcomes != 1 || r.Status != "ARCHIVE_REVIEW_REQUIRED" {
		t.Fatalf("unexpected: %+v", r)
	}
	if x := AuditIntegrity(nil, now); x.Status != "NO_OBSERVATIONS" {
		t.Fatal(x)
	}
}
