package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"testing"
	"time"
)

func TestReviewMissingNeverManufacturesPerformance(t *testing.T) {
	cfg := paper.DefaultConfig()
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	exit := now.Add(time.Minute)
	p := paper.Position{Market: "btc-updown-5m-111", Side: "UP", ConditionID: "a", EntryTime: now, MarketEnd: now.Add(5 * time.Minute), ExitTime: &exit, EntryCost: 20, EntryPrice: .502, Qty: 40, RealizedPnL: -5, Status: "CLOSED"}
	r := BuildReview([]paper.Position{p}, cfg, Archive{})
	if r.TradingAuthorized || r.MatchedPairs != 0 || r.Status != "NO_RECORDED_ORDERBOOK_ARCHIVE" || r.AllRecordedClosed != 1 || r.MissingOrUnverified != 1 {
		t.Fatalf("missing should not be scored %+v", r)
	}
	frame := func(t time.Time, bid float64) Frame {
		return Frame{Time: t, BTC: 95000, Source: "Coinbase", Markets: []polymarket.Market{{Slug: p.Market, ConditionID: "a", Active: true, Up: polymarket.Quote{Bid: bid, Ask: bid + .01, BidDepth: 100, AskDepth: 100, UpdatedAt: t}}}}
	}
	f1 := frame(now, .49)
	f2 := frame(now.Add(5*time.Second), .58)
	r = BuildReview([]paper.Position{p}, cfg, Archive{Frames: []Frame{f1, f2}, Files: 1})
	if r.MatchedPairs != 1 || r.Groups[0].PairedModeledExits != 1 || r.Groups[0].IndependentUTCDays != 1 || r.Status != "INSUFFICIENT_INDEPENDENT_REPLAY_EVIDENCE" {
		t.Fatalf("paired replay invalid %+v", r)
	}
	if r.Groups[0].RecordedPnLOnPairsUSD != -5 {
		t.Fatal("recorded realized profit overwritten")
	}
}
