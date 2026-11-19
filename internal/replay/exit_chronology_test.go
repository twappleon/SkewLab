package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"testing"
	"time"
)

func TestReplayCannotExitAtEntryOrUseFutureQuote(t *testing.T) {
	t0 := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	p := paper.Position{Market: "m", ConditionID: "c", Side: "UP", EntryTime: t0, MarketEnd: t0.Add(5 * time.Minute), EntryCost: 20, Qty: 40}
	f := func(at, quoteAt time.Time) Frame {
		return Frame{Time: at, Markets: []polymarket.Market{{Slug: "m", ConditionID: "c", Active: true, Up: polymarket.Quote{Bid: .1, Ask: .11, BidDepth: 200, AskDepth: 200, UpdatedAt: quoteAt}}}}
	}
	cfg := paper.DefaultConfig()
	r := ReplayExit(p, []Frame{f(t0, t0)}, cfg, OriginalPolicy)
	if r.ModeledPnLUSD != nil {
		t.Fatalf("same instant order and exit: %+v", r)
	}
	r = ReplayExit(p, []Frame{f(t0.Add(6*time.Second), t0.Add(7*time.Second))}, cfg, OriginalPolicy)
	if r.ModeledPnLUSD != nil {
		t.Fatalf("future quote used for stop: %+v", r)
	}
	r = ReplayExit(p, []Frame{f(t0.Add(6*time.Second), t0.Add(6*time.Second))}, cfg, OriginalPolicy)
	if r.Reason != "stop_loss" || r.ModeledPnLUSD == nil {
		t.Fatalf("valid stop not counted: %+v", r)
	}
}
