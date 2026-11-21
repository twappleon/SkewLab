package research100

import (
	"testing"
	"time"
)

func TestCoverageNoFabricatedDenominator(t *testing.T) {
	now := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	d, e := BuildDataset([]Observation{fixture(now.Add(30 * time.Second))})
	if e != nil {
		t.Fatal(e)
	}
	c := AssessCoverage(d, 5*time.Minute, nil, nil)
	if c.Expected != nil || c.Fraction != nil {
		t.Fatal("invented coverage")
	}
	end := now.Add(15 * time.Minute)
	c = AssessCoverage(d, 5*time.Minute, &now, &end)
	if c.Expected == nil || *c.Expected != 3 || c.Observed != 1 || c.Missing != 2 {
		t.Fatalf("bad coverage: %+v", c)
	}
}
