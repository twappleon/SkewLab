package main

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"testing"
	"time"
)

func TestSummarizeRejectsDemoCountsRealDaysAndGaps(t *testing.T) {
	now := time.Date(2026, 11, 18, 23, 59, 58, 0, time.UTC)
	m := polymarket.Market{Active: true, Up: polymarket.Quote{Bid: .44, Ask: .45, BidDepth: 10, AskDepth: 10, UpdatedAt: now}, Down: polymarket.Quote{Bid: .54, Ask: .55, BidDepth: 10, AskDepth: 10, UpdatedAt: now}}
	o := []Observation{
		{Time: now, BTC: 98000, Source: "Binance BTCUSDT spot", Markets: []polymarket.Market{m}},
		{Time: now.Add(time.Second), BTC: 98001, Source: "SYNTHETIC DEMO (not live)"},
		{Time: now.Add(15 * time.Second), BTC: 98005, Source: "Binance BTCUSDT spot", Markets: []polymarket.Market{m}},
	}
	r := summarize(o)
	if r.Records != 2 || r.Invalid != 1 || r.LargeGaps != 1 || r.CalendarDays != 2 || r.FreshPairs != 1 {
		t.Fatalf("bad report: %+v", r)
	}
}
