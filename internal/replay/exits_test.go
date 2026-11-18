package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"testing"
	"time"
)

func TestReplayExitRequiresRealFreshBidAndDepth(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	cfg := paper.DefaultConfig()
	p := paper.Position{Market: "btc-updown-5m-111", ConditionID: "x", Side: "UP", EntryTime: now, MarketEnd: now.Add(5 * time.Minute), EntryCost: 20, Qty: 40}
	mk := func(at time.Time, bid, depth float64) Frame {
		return Frame{Time: at, Markets: []polymarket.Market{{Slug: p.Market, ConditionID: "x", Active: true, Up: polymarket.Quote{Bid: bid, Ask: bid + .01, BidDepth: depth, AskDepth: 100, UpdatedAt: at}}}}
	}
	good := mk(now.Add(6*time.Second), .58, 100)
	cases := []Frame{mk(now.Add(time.Second), .12, 100), mk(now.Add(3*time.Second), .12, 100), good}
	r := ReplayExit(p, cases, cfg, OriginalPolicy)
	if r.Status != "MODELED_EXIT_FROM_RECORDED_QUOTE" || r.Reason != "take_profit" || r.ModeledPnLUSD == nil || *r.ModeledPnLUSD <= 0 {
		t.Fatalf("no valid replay %+v", r)
	}
	r = ReplayExit(p, []Frame{mk(now.Add(6*time.Second), .30, 10)}, cfg, OriginalPolicy)
	if r.Status != "NO_VERIFIABLE_EXIT" || r.ModeledPnLUSD != nil {
		t.Fatal("invented fill without depth")
	}
	stale := mk(now.Add(6*time.Second), .58, 100)
	stale.Markets[0].Up.UpdatedAt = now.Add(-time.Minute)
	r = ReplayExit(p, []Frame{stale}, cfg, OriginalPolicy)
	if r.ModeledPnLUSD != nil {
		t.Fatal("used stale frame")
	}
	r = ReplayExit(p, []Frame{mk(now.Add(6*time.Minute), .58, 100)}, cfg, OriginalPolicy)
	if r.ModeledPnLUSD != nil {
		t.Fatal("used post-expiry frame")
	}
}
func TestReplayStopPolicyIsNotPastPerformance(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	cfg := paper.DefaultConfig()
	p := paper.Position{Market: "m", ConditionID: "id", Side: "UP", EntryTime: now, MarketEnd: now.Add(5 * time.Minute), EntryCost: 20, Qty: 40}
	f := Frame{Time: now.Add(5 * time.Second), Markets: []polymarket.Market{{Slug: "m", ConditionID: "id", Active: true, Up: polymarket.Quote{Bid: .49, Ask: .50, AskDepth: 80, BidDepth: 80, UpdatedAt: now.Add(5 * time.Second)}}}}
	a := ReplayExit(p, []Frame{f}, cfg, OriginalPolicy)
	b := ReplayExit(p, []Frame{f}, cfg, AlternativePolicy)
	if a.Reason != "stop_loss" || b.ModeledPnLUSD != nil || b.Status != "NO_VERIFIABLE_EXIT" {
		t.Fatalf("policies not distinct: %+v %+v", a, b)
	}
}
