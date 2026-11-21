package research100

import (
	"math"
	"testing"
	"time"
)

func fixture(t time.Time) Observation {
	return Observation{ID: "m1", Market: "5m", Side: "UP", CapturedAt: t, ExpiresAt: t.Add(time.Minute), QuoteAt: t.Add(-time.Second), Bid: .49, Ask: .51, BidDepth: 10, AskDepth: 10, Momentum: .04, Skew: .3}
}
func TestObservationRejectsFutureAndMalformedEvidence(t *testing.T) {
	now := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	v := fixture(now)
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	c := v
	c.QuoteAt = now.Add(time.Second)
	if c.Validate() == nil {
		t.Fatal("future book accepted")
	}
	c = v
	c.Bid = math.NaN()
	if c.Validate() == nil {
		t.Fatal("NaN accepted")
	}
	c = v
	c.Winner = "UP"
	c.ResolvedAt = now
	if c.Validate() == nil {
		t.Fatal("lookahead accepted")
	}
	c = v
	c.Winner = "UP"
	c.ResolvedAt = c.ExpiresAt.Add(time.Second)
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}
