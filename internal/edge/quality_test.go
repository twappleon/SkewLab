package edge

import (
	"testing"
	"time"

	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
)

func TestAssessQualityOnlyObservedMarkets(t *testing.T) {
	dir := t.TempDir()
	r := NewRecorder(dir)
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	m := polymarket.Market{Slug: "btc-updown-5m-100", ConditionID: "id5", Seconds: 300, End: now.Add(4 * time.Minute), Active: true, Up: polymarket.Quote{Ask: .45, Bid: .43, AskDepth: 40, BidDepth: 40, UpdatedAt: now}}
	s := strategy.Result{Direction: "UP", Ready: true, MomentumPct: .04, Skew: .4}
	captured, err := r.Capture(m, s, now)
	if err != nil || !captured {
		t.Fatalf("capture %v %v", captured, err)
	}
	q := r.AssessQuality(now)
	if q.Status != "COLLECTING_INSUFFICIENT_EVIDENCE" || q.Total != 1 || q.Pending != 1 || q.Overdue != 0 || q.Groups[0].Captured != 1 || q.TradingAuthorized {
		t.Fatalf("bad initial quality %+v", q)
	}
	overdue := m.End.Add(31 * time.Minute)
	q = r.AssessQuality(overdue)
	if q.Status != "SETTLEMENT_BACKLOG_REVIEW" || q.Overdue != 1 || q.Groups[0].Overdue != 1 {
		t.Fatalf("not overdue %+v", q)
	}
	if err := r.Resolve(m.Slug, m.ConditionID, "DOWN", overdue); err != nil {
		t.Fatal(err)
	}
	q = r.AssessQuality(overdue.Add(time.Second))
	if q.Overdue != 0 || q.Resolved != 1 || q.Pending != 0 || q.LargestCohort != 1 || q.LastResolutionAt == nil || q.TradingAuthorized {
		t.Fatalf("bad resolved quality %+v", q)
	}
	restored := NewRecorder(dir)
	if restored.AssessQuality(overdue.Add(2*time.Second)).Resolved != 1 {
		t.Fatal("quality lost after restart")
	}
}
func TestAssessQualityNeverClaimsTradingReadiness(t *testing.T) {
	now := time.Now().UTC()
	r := NewRecorder(t.TempDir())
	r.mu.Lock()
	for i := 0; i < MinEvidence; i++ {
		tm := now.Add(-time.Duration(i+200) * 10 * time.Minute)
		r.records = append(r.records, Candidate{Slug: "test", ConditionID: "condition", WindowSeconds: 300, Side: "UP", Ask: .45, CapturedAt: tm, MarketEnd: tm.Add(5 * time.Minute), Winner: "UP", ResolvedAt: tm.Add(6 * time.Minute)})
	}
	r.mu.Unlock()
	q := r.AssessQuality(now)
	if q.LargestCohort != MinEvidence || q.Status != "NO_RECENT_QUALIFYING_SIGNALS" || q.TradingAuthorized {
		t.Fatalf("incorrect readiness: %+v", q)
	}
	// Recent candidates still never become a trading permission.
	r.mu.Lock()
	r.records[0].CapturedAt = now.Add(-time.Hour)
	r.records[0].MarketEnd = now.Add(-55 * time.Minute)
	r.records[0].ResolvedAt = now.Add(-54 * time.Minute)
	r.mu.Unlock()
	q = r.AssessQuality(now)
	if q.Status != "RESEARCH_REVIEW_ONLY" || q.TradingAuthorized {
		t.Fatalf("evidence alone incorrectly authorizes trading: %+v", q)
	}
}
func TestAssessQualityInvalidRecordsFlagged(t *testing.T) {
	now := time.Now().UTC()
	r := NewRecorder(t.TempDir())
	r.records = []Candidate{{CapturedAt: now.Add(-time.Hour), MarketEnd: now.Add(-time.Minute), Winner: "UP", ResolvedAt: now.Add(-time.Hour)}}
	q := r.AssessQuality(now)
	if q.Status != "INVALID_RESEARCH_RECORDS" || q.Invalid != 1 || q.Resolved != 0 {
		t.Fatalf("invalid record counted: %+v", q)
	}
}
