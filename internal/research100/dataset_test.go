package research100

import (
	"testing"
	"time"
)

func TestDatasetRejectDuplicatesWithoutInflatingSamples(t *testing.T) {
	at := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	a := fixture(at)
	if _, err := BuildDataset([]Observation{a, a}); err == nil {
		t.Fatal("duplicate market accepted")
	}
	b := fixture(at.Add(time.Minute))
	b.ID = "different"
	d, e := BuildDataset([]Observation{b, a})
	if e != nil {
		t.Fatal(e)
	}
	if d.Items[0].ID != "m1" || len(d.Items) != 2 {
		t.Fatal("bad deterministic order")
	}
	a.Winner = "UP"
	a.ResolvedAt = at.Add(time.Second)
	if _, err := BuildDataset([]Observation{a}); err == nil {
		t.Fatal("future label accepted")
	}
}
