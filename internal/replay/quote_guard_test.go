package replay

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"math"
	"testing"
	"time"
)

func TestQuoteGuardRejectsMalformedBooks(t *testing.T) {
	n := time.Date(2026, 11, 19, 1, 0, 0, 0, time.UTC)
	good := polymarket.Quote{Bid: .49, Ask: .5, BidDepth: 100, AskDepth: 100, UpdatedAt: n.Add(-time.Second)}
	if !validQuote(good, n) {
		t.Fatal("good book rejected")
	}
	cases := []polymarket.Quote{}
	add := func(f func(*polymarket.Quote)) { x := good; f(&x); cases = append(cases, x) }
	add(func(q *polymarket.Quote) { q.Bid = 0 })
	add(func(q *polymarket.Quote) { q.Ask = 0 })
	add(func(q *polymarket.Quote) { q.Bid = .8 })
	add(func(q *polymarket.Quote) { q.Ask = 1 })
	add(func(q *polymarket.Quote) { q.BidDepth = 0 })
	add(func(q *polymarket.Quote) { q.AskDepth = 0 })
	add(func(q *polymarket.Quote) { q.Ask = math.NaN() })
	add(func(q *polymarket.Quote) { q.BidDepth = math.Inf(1) })
	add(func(q *polymarket.Quote) { q.UpdatedAt = n.Add(time.Millisecond) })
	add(func(q *polymarket.Quote) { q.UpdatedAt = n.Add(-13 * time.Second) })
	for i, q := range cases {
		if validQuote(q, n) {
			t.Fatalf("invalid case %d accepted: %+v", i, q)
		}
	}
}
