package main

import (
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/research100"
	"testing"
	"time"
)

func TestResearchToolCannotAuthorizeTrading(t *testing.T) {
	now := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	a := research100.Observation{ID: "m1", Market: "5m", Side: "UP", CapturedAt: now, ExpiresAt: now.Add(time.Minute), QuoteAt: now.Add(-time.Second), Bid: .49, Ask: .51, BidDepth: 10, AskDepth: 10, Momentum: .03, Skew: .3}
	b, e := json.Marshal([]research100.Observation{a})
	if e != nil {
		t.Fatal(e)
	}
	out, e := Execute("audit", b, now.Add(time.Hour))
	if e != nil {
		t.Fatal(e)
	}
	if out.TradingAuthorized || out.WalkForward == nil || out.WalkForward.ScoredResolved != 0 || out.Coverage.Fraction != nil {
		t.Fatalf("false evidence: %+v", out)
	}
	cmp, e := Execute("comparison", []byte("[]"), now)
	if e != nil || cmp.Comparison == nil || cmp.TradingAuthorized {
		t.Fatalf("bad no-data comparison %+v %v", cmp, e)
	}
	if _, e = Execute("audit", []byte("broken"), now); e == nil {
		t.Fatal("invalid archive accepted")
	}
}
