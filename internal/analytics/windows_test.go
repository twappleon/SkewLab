package analytics

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"testing"
	"time"
)

func TestCompareMarketsSeparateAndExitReason(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	day := now.Add(-time.Hour)
	p := func(market, status, note string, value float64) paper.Position {
		return paper.Position{Market: market, Status: status, Note: note, ExitTime: &day, RealizedPnL: value}
	}
	l := paper.New(paper.DefaultConfig())
	l.Positions = []paper.Position{
		p("btc-updown-5m-1795000000", "CLOSED", "net take profit", 2),
		p("btc-updown-5m-1795000300", "SETTLED", "official result", -18),
		p("btc-updown-15m-1795000000", "CLOSED", "net stop loss", -1),
		p("othermarket", "CLOSED", "exit before settlement", 3),
		{Market: "btc-updown-15m-x", Status: "PENDING_RESOLUTION", ExitTime: &day, RealizedPnL: 5000},
	}
	out := CompareMarkets(l, now)
	if len(out) != 3 {
		t.Fatalf("three groups required: %+v", out)
	}
	if out[0].Period.Closed != 2 || out[0].Period.NetPnLUSD != -16 || out[0].Exits.TakeProfit != 1 || out[0].Exits.Settled != 1 {
		t.Fatalf("5m: %+v", out[0])
	}
	if out[0].Period.ProfitFactor == nil || *out[0].Period.ProfitFactor != 0.1111 {
		t.Fatalf("5m profit factor: %+v", out[0].Period)
	}
	if out[1].Period.Closed != 1 || out[1].Period.NetPnLUSD != -1 || out[1].Exits.StopLoss != 1 {
		t.Fatalf("15m: %+v", out[1])
	}
	if out[1].Period.ProfitFactor == nil || *out[1].Period.ProfitFactor != 0 {
		t.Fatalf("15m PF 0 with loss: %+v", out[1].Period)
	}
	if out[2].Period.Closed != 1 || out[2].Exits.BeforeSettlement != 1 {
		t.Fatalf("other: %+v", out[2])
	}
	for _, w := range out {
		if w.SampleLevel != "insufficient" {
			t.Fatalf("small samples must be insufficient: %+v", w)
		}
	}
}

func TestCompareMarketsEmptyAndCalendarEdges(t *testing.T) {
	now := time.Date(2026, 11, 18, 6, 0, 0, 0, time.UTC)
	start := time.Date(2026, 11, 4, 17, 0, 0, 0, time.UTC) // Nov 5 midnight in Phnom Penh: included
	before := start.Add(-time.Nanosecond)
	future := now.Add(time.Second)
	l := paper.New(paper.DefaultConfig())
	l.Positions = []paper.Position{
		{Market: "btc-updown-5m-1", Status: "CLOSED", ExitTime: &start, RealizedPnL: 1},
		{Market: "btc-updown-5m-2", Status: "CLOSED", ExitTime: &before, RealizedPnL: 500},
		{Market: "btc-updown-5m-3", Status: "CLOSED", ExitTime: &future, RealizedPnL: 500},
	}
	out := CompareMarkets(l, now)
	if out[0].Period.Closed != 1 || out[0].Period.NetPnLUSD != 1 {
		t.Fatalf("bad time window: %+v", out[0])
	}
	if out[0].Period.ProfitFactor != nil {
		t.Fatal("profit factor without any losses should be nil")
	}
	if out[1].Period.Closed != 0 || out[1].Period.WinRatePct != nil {
		t.Fatalf("empty window: %+v", out[1])
	}
	if CompareMarkets(nil, now)[0].Period.Closed != 0 {
		t.Fatal("nil ledger should not fail")
	}
}
