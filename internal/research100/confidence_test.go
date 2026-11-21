package research100

import (
	"testing"
	"time"
)

func TestEdgeRequiresPriorOutcomesAcrossDays(t *testing.T) {
	anchor := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	items := make([]Observation, 110)
	for i := range items {
		at := anchor.Add(time.Duration(i/10)*24*time.Hour + time.Duration(i%10)*5*time.Minute)
		v := fixture(at)
		v.ID = at.Format(time.RFC3339Nano)
		v.Winner = "UP"
		v.ResolvedAt = v.ExpiresAt.Add(time.Minute)
		items[i] = v
	}
	d, e := BuildDataset(items)
	if e != nil {
		t.Fatal(e)
	}
	now := anchor.Add(13 * 24 * time.Hour)
	res := EvaluateEdge(d, "5m", "UP", now, .51, .002, 25)
	if res.Status != "RESEARCH_CANDIDATE_ONLY" || res.TradingAuthorized || res.UTCObservationDays != 11 {
		t.Fatalf("bad research proof %+v", res)
	}
	early := EvaluateEdge(d, "5m", "UP", anchor.Add(5*24*time.Hour), .51, .002, 25)
	if early.Status != "INSUFFICIENT_PRIOR_EVIDENCE" || early.Samples >= 100 {
		t.Fatal("future labels leaked")
	}
}
