package analytics

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"math"
	"testing"
	"time"
)

func TestDiagnosticStopOvershootAndAccounting(t *testing.T) {
	now := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	entry := now.Add(-time.Minute)
	p := paper.Position{Market: "btc-updown-5m-123", Status: "CLOSED", Note: "net stop loss", EntryTime: entry, ExitTime: &now, Qty: 20, EntryPrice: .51, ExitPrice: .02, EntryCost: 10.4, RealizedPnL: -10.06}
	d, ok := ExplainCompleted(p, 5)
	if !ok || !d.AttributionComplete || d.MarketGroup != "5m" || !d.ExceededStopThreshold {
		t.Fatalf("wrong diagnostic: %+v ok %v", d, ok)
	}
	if d.LossBeyondStopUSD != 9.54 || *d.ContractPricePnLUSD != -9.8 || *d.ModeledCostReserveUSD != .26 {
		t.Fatalf("unexpected accounting %+v", d)
	}
	if d.HoldingSeconds == nil || *d.HoldingSeconds != 60 {
		t.Fatalf("holding period %+v", d)
	}
	l := paper.New(paper.DefaultConfig())
	l.Positions = []paper.Position{p}
	l.TotalClosed = 1
	s := SummarizeCauses(l, now)
	if s.CompletedTrades != 1 || s.StopThresholdBreaches != 1 || s.MissingAttribution != 0 || len(s.ByReasonMarket) != 1 {
		t.Fatalf("bad summary %+v", s)
	}
	if math.Abs((s.PriceMovementPnLUSD-s.CostReserveUSD)-s.NetPnLUSD) > .0001 {
		t.Fatalf("bad identity %+v", s)
	}
}
func TestDiagnosticsSettlementAndInvalidFields(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	t0 := now.Add(-time.Minute)
	settled := paper.Position{Market: "btc-updown-15m-123", Status: "SETTLED", Note: "official result", EntryTime: t0, ExitTime: &now, Qty: 20, EntryPrice: .6, ExitPrice: 1, EntryCost: 12.5, RealizedPnL: 7.5}
	d, ok := ExplainCompleted(settled, 5)
	if !ok || !d.AttributionComplete || *d.ContractPricePnLUSD != 8 || *d.ModeledCostReserveUSD != .5 || d.ExitReason != "official_settlement" {
		t.Fatalf("settlement: %+v", d)
	}
	pos := paper.Position{Status: "PENDING_RESOLUTION", ExitTime: &now, EntryCost: 20, RealizedPnL: 10000}
	l := paper.New(paper.DefaultConfig())
	l.Positions = []paper.Position{settled, pos}
	summary := SummarizeCauses(l, now)
	if summary.CompletedTrades != 1 || summary.NetPnLUSD != 7.5 || summary.StopThresholdBreaches != 0 {
		t.Fatal(summary)
	}
	broken := settled
	broken.Qty = 0
	d, ok = ExplainCompleted(broken, 5)
	if !ok || d.AttributionComplete {
		t.Fatal("incomplete attribution must be clear")
	}
	future := settled
	future.ExitTime = new(time.Time)
	*future.ExitTime = now.Add(time.Hour)
	l.Positions = append(l.Positions, broken, future)
	summary = SummarizeCauses(l, now)
	if summary.CompletedTrades != 2 || summary.MissingAttribution != 1 {
		t.Fatalf("future or missing record handled incorrectly %+v", summary)
	}
}
func TestDiagnosticsNoNaNNoPrivateLeak(t *testing.T) {
	now := time.Now().UTC()
	l := paper.New(paper.DefaultConfig())
	l.Positions = []paper.Position{{ID: "secret-trade-id", Market: "secret-market-id", Token: "secret-token", Status: "CLOSED", ExitTime: &now, EntryCost: 20, RealizedPnL: math.NaN()}}
	if SummarizeCauses(l, now).CompletedTrades != 0 {
		t.Fatal("NaN must be rejected")
	}
	if PrivateDiagnostics(l, now, 100) == nil {
		t.Fatal("must return [] even when empty")
	}
}
