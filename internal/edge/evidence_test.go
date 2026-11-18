package edge

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func market(now time.Time) polymarket.Market {
	return polymarket.Market{Slug: "btc-updown-5m-999", ConditionID: "condition-test", Seconds: 300, End: now.Add(4 * time.Minute), Active: true,
		Up: polymarket.Quote{Ask: .40, Bid: .39, AskDepth: 300, BidDepth: 300, UpdatedAt: now}, Down: polymarket.Quote{Ask: .60, Bid: .59, AskDepth: 300, BidDepth: 300, UpdatedAt: now}}
}
func signal() strategy.Result {
	return strategy.Result{Direction: "UP", Ready: true, MomentumPct: .04, Skew: .4}
}
func TestPersistentFirstSignalAndVerifiedOnly(t *testing.T) {
	dir := t.TempDir()
	r := NewRecorder(dir)
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	m := market(now)
	if yes, err := r.Capture(m, signal(), now); err != nil || !yes {
		t.Fatalf("capture: %v, %v", yes, err)
	}
	if yes, err := r.Capture(m, signal(), now); err != nil || yes {
		t.Fatalf("duplicate: %v, %v", yes, err)
	}
	if len(r.Due(now, 10)) != 0 {
		t.Fatal("must not query outcome before expiry")
	}
	restored := NewRecorder(dir)
	if restored.Status().Total != 1 {
		t.Fatalf("restart: %+v", restored.Status())
	}
	if yes, _ := restored.Capture(m, signal(), now); yes {
		t.Fatal("duplicate after restart")
	}
	due := restored.Due(m.End.Add(31*time.Second), 10)
	if len(due) != 1 || due[0].ConditionID != "condition-test" {
		t.Fatal("due result missing identity")
	}
	if err := restored.Resolve(m.Slug, "wrong", "UP", m.End.Add(time.Minute)); err == nil {
		t.Fatal("wrong condition accepted")
	}
	if err := restored.Resolve(m.Slug, m.ConditionID, "UP", m.End.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := restored.Resolve(m.Slug, m.ConditionID, "UP", m.End.Add(time.Minute)); err != nil {
		t.Fatal("non-idempotent result: ", err)
	}
	if err := restored.Resolve(m.Slug, m.ConditionID, "DOWN", m.End.Add(time.Minute)); err == nil {
		t.Fatal("contradiction accepted")
	}
	if restored.Status().Resolved != 1 || len(restored.Due(m.End.Add(time.Hour), 10)) != 0 {
		t.Fatal("resolution repeated")
	}
	reread := NewRecorder(dir)
	if reread.Status().Resolved != 1 {
		t.Fatal("resolution not persistent")
	}
}
func TestNoFabricationOnBadInputsAndCorruptFile(t *testing.T) {
	dir := t.TempDir()
	r := NewRecorder(dir)
	now := time.Now().UTC()
	m := market(now)
	m.Up.UpdatedAt = now.Add(-13 * time.Second)
	if yes, err := r.Capture(m, signal(), now); yes || err != nil {
		t.Fatalf("stale book accepted: %v, %v", yes, err)
	}
	m.Up.UpdatedAt = now
	m.Active = false
	if yes, _ := r.Capture(m, signal(), now); yes {
		t.Fatal("inactive market accepted")
	}
	if err := os.WriteFile(filepath.Join(dir, "edge-research.json"), []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	bad := NewRecorder(dir)
	if bad.Status().LastError == "" {
		t.Fatal("invalid historical journal not flagged")
	}
	if yes, err := bad.Capture(m, signal(), now); yes || err == nil {
		t.Fatal("invalid history overwritten")
	}
}
func TestConfidenceRequiresPriorIndependentOutcomes(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	history := make([]Candidate, 0, 101)
	for i := 0; i < 100; i++ {
		tm := now.Add(-time.Duration(103-i) * 5 * time.Minute)
		history = append(history, Candidate{Slug: "unique" + time.Unix(int64(i), 0).Format("150405"), ConditionID: "c", WindowSeconds: 300, Side: "UP", CapturedAt: tm, MarketEnd: tm.Add(4 * time.Minute), Ask: .4, Winner: "UP", ResolvedAt: tm.Add(6 * time.Minute)})
	}
	decision := EvaluatePast(history, now, 300, "UP", .40, .002, 25)
	if decision.Status != "RESEARCH_CANDIDATE_ONLY" || decision.Samples != 100 || decision.WilsonLower95 == nil || *decision.WilsonLower95 >= 1 {
		t.Fatalf("confidence not lower bounded: %+v", decision)
	}
	decision = EvaluatePast(history[:99], now, 300, "UP", .40, .002, 25)
	if decision.Status != "NO_EDGE_EVIDENCE" {
		t.Fatal("insufficient sample incorrectly called edge")
	}
	decision = EvaluatePast(history, now, 900, "UP", .40, .002, 25)
	if decision.Samples != 0 {
		t.Fatal("5m results leaked into 15m cohort")
	}
	decision = EvaluatePast(history, now, 300, "UP", .8, .002, 25)
	if decision.Samples != 0 {
		t.Fatal("price cohort leakage")
	}
	history[0].ResolvedAt = now.Add(time.Hour)
	decision = EvaluatePast(history, now, 300, "UP", .40, .002, 25)
	if decision.Samples != 99 {
		t.Fatal("future resolution included")
	}
}
func TestWalkForwardRejectsLookahead(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	records := make([]Candidate, 0, 130)
	for i := 0; i < 120; i++ {
		tm := now.Add(-time.Duration(121-i) * 10 * time.Minute)
		records = append(records, Candidate{Slug: "sample-" + time.Unix(int64(i), 0).Format("150405"), WindowSeconds: 300, Side: "UP", Ask: .4, CapturedAt: tm, MarketEnd: tm.Add(5 * time.Minute), ResolvedAt: tm.Add(6 * time.Minute), Winner: "UP"})
	}
	result := WalkForwardCheck(records, .002, 25)
	if result.Resolved != 120 || result.EligibleResearchCandidates != 20 {
		t.Fatalf("walkforward counts: %+v", result)
	}
	// If every outcome becomes known only after all prospective trades,
	// the same 120 favorable labels provide ZERO available prior evidence.
	for i := range records {
		records[i].ResolvedAt = now.Add(time.Hour)
	}
	result = WalkForwardCheck(records, .002, 25)
	if result.EligibleResearchCandidates != 0 {
		t.Fatalf("look-ahead in walkforward: %+v", result)
	}
}
func TestConservativeCostAndNoNaN(t *testing.T) {
	d := EvaluatePast(nil, time.Now(), 300, "UP", math.NaN(), 0, 25)
	if d.Status != "INVALID_INPUT" {
		t.Fatal("non-finite price accepted")
	}
	if math.Abs(wilsonLower(50, 100)-.404) < .01 {
	} else {
		t.Fatalf("unexpected Wilson bound: %v", wilsonLower(50, 100))
	}
}
