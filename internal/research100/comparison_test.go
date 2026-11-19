package research100

import (
	"testing"
	"time"
)

func TestRulesUseSameSnapshotPoolWithAbstentions(t *testing.T) {
	at := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	a := fixture(at)
	a.Winner = "UP"
	a.ResolvedAt = a.ExpiresAt.Add(time.Second)
	b := fixture(at.Add(5 * time.Minute))
	b.ID = "m2"
	b.Winner = "DOWN"
	b.ResolvedAt = b.ExpiresAt.Add(time.Second)
	b.Skew = -.1
	d, e := BuildDataset([]Observation{a, b})
	if e != nil {
		t.Fatal(e)
	}
	p := CompareRules(d, at.Add(time.Hour))
	if p.OutcomeLabeled != 2 || p.Rules[0].Decisions != 2 || p.Rules[2].Decisions != 1 || p.Rules[2].EvaluatedSameSnapshots != 2 {
		t.Fatalf("bad paired pool %+v", p)
	}
}
