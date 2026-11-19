package research100

import (
	"testing"
	"time"
)

func TestWalkForwardCausalFrontierAndNoTradeApproval(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	raw := make([]Observation, 130)
	for i := range raw {
		at := start.Add(time.Duration(i/10)*24*time.Hour + time.Duration(i%10)*5*time.Minute)
		v := fixture(at)
		v.ID = at.Format(time.RFC3339Nano)
		v.Winner = "UP"
		v.ResolvedAt = v.ExpiresAt.Add(time.Minute)
		raw[i] = v
	}
	d, e := BuildDataset(raw)
	if e != nil {
		t.Fatal(e)
	}
	w := RunWalkForward(d, start.Add(15*24*time.Hour))
	if w.EligibleResearchCandidates < 25 || w.ScoredResolved != w.EligibleResearchCandidates || w.TradingAuthorized {
		t.Fatalf("bad forward %+v", w)
	}
	for i := range raw {
		raw[i].ResolvedAt = start.Add(30 * 24 * time.Hour)
	}
	d, e = BuildDataset(raw)
	if e != nil {
		t.Fatal(e)
	}
	w = RunWalkForward(d, start.Add(16*24*time.Hour))
	if w.EligibleResearchCandidates != 0 {
		t.Fatalf("leaked future labels %+v", w)
	}
}
