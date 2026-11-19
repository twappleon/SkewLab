package research100

import (
	"math"
	"testing"
	"time"
)

func TestCalibrationOnlyResolvedPast(t *testing.T) {
	at := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	a := fixture(at)
	a.Winner = "UP"
	a.ResolvedAt = at.Add(time.Minute)
	b := fixture(at.Add(2 * time.Minute))
	b.ID = "m2"
	b.Winner = "DOWN"
	b.ResolvedAt = b.ExpiresAt.Add(time.Second)
	d, e := BuildDataset([]Observation{a, b})
	if e != nil {
		t.Fatal(e)
	}
	s := ScoreCalibration(d, at.Add(90*time.Second))
	if s.Samples != 1 || s.Brier == nil || math.Abs(*s.Brier-.25) > 1e-6 {
		t.Fatalf("bad asof %+v", s)
	}
	s = ScoreCalibration(d, at.Add(4*time.Minute))
	if s.Samples != 2 || s.LogLoss == nil || s.BaseRate == nil {
		t.Fatalf("bad score %+v", s)
	}
}
