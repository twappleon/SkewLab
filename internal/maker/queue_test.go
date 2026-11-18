package maker

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/stream"
	"testing"
	"time"
)

func TestQueueNeverBooksAsPnL(t *testing.T) {
	q := NewQueueMonitor()
	now := time.Now().UTC()
	m := polymarket.Market{Slug: "btc-updown-5m-1", Active: true, End: now.Add(5 * time.Minute), UpToken: "up", Up: polymarket.Quote{Bid: .4, BidDepth: 5}}
	proposal := Proposal{Status: "INDICATIVE_ONLY_NO_FILLS", Side: "UP", SizeUSD: 4}
	q.Update(m, proposal, now)
	if len(q.Snapshot()) != 1 {
		t.Fatal("missing queue research")
	}
	// Trade before study starts can never fill hypothetical order.
	if _, ok := q.Observe(stream.Trade{Token: "up", Price: .4, Side: "SELL", Size: 100, Time: now.Add(-time.Second)}); ok {
		t.Fatal("lookahead")
	}
	// Public trade volume must consume visible queue plus hypothetical quantity.
	v, ok := q.Observe(stream.Trade{Token: "up", Price: .4, Side: "SELL", Size: 5, Time: now.Add(time.Second)})
	if !ok || v.Candidate {
		t.Fatal("premature fill candidate")
	}
	_, ok = q.Observe(stream.Trade{Token: "up", Price: .4, Side: "SELL", Size: 100, Time: now.Add(time.Second)})
	if ok {
		t.Fatal("same timestamp duplicate counted")
	}
	v, ok = q.Observe(stream.Trade{Token: "up", Price: .4, Side: "SELL", Size: 10, Time: now.Add(2 * time.Second)})
	if !ok || !v.Candidate || q.Candidates() != 1 {
		t.Fatal("no fill research candidate")
	}
	if v.Status != "HYPOTHETICAL_QUEUE_CLEARED_NOT_ACTUAL_FILL" {
		t.Fatal("must not claim real fill")
	}
}
