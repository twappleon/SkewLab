package paper

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
	"testing"
	"time"
)

func TestPaperEntryExitAndFees(t *testing.T) {
	now := time.Unix(1800000000, 0)
	c := DefaultConfig()
	c.CooldownSeconds = 0
	l := New(c)
	m := polymarket.Market{Slug: "test", End: now.Add(4 * time.Minute), Active: true, UpToken: "up", DownToken: "down", Up: polymarket.Quote{Bid: .47, Ask: .48, AskDepth: 1000, BidDepth: 1000}}
	s := strategy.Result{Direction: "UP", Ready: true}
	msgs := l.Evaluate(m, s, now)
	if len(msgs) != 1 || l.Cash >= 1000 {
		t.Fatalf("entry failed %+v", l)
	}
	m.Up.Bid = .55
	l.Evaluate(m, strategy.Result{Direction: "WAIT"}, now.Add(5*time.Second))
	if l.TotalClosed != 1 || l.RealizedPnL <= 0 || l.Cash <= 1000 {
		t.Fatalf("exit failed %+v", l)
	}
}
func TestNoEntryWithoutDepthOrNearExpiry(t *testing.T) {
	now := time.Unix(1800000000, 0)
	l := New(DefaultConfig())
	s := strategy.Result{Direction: "UP", Ready: true}
	m := polymarket.Market{Slug: "m", Active: true, End: now.Add(2 * time.Minute), Up: polymarket.Quote{Ask: .5, AskDepth: 2}}
	l.Evaluate(m, s, now)
	if len(l.Positions) != 0 {
		t.Fatal("entered without liquidity")
	}
	m.Up.AskDepth = 1000
	m.End = now.Add(45 * time.Second)
	l.Evaluate(m, s, now)
	if len(l.Positions) != 0 {
		t.Fatal("entered near expiry")
	}
}

func TestSettlementIdempotenceAndPendingConservativeValuation(t *testing.T) {
	now := time.Unix(1800000000, 0)
	c := DefaultConfig()
	c.CooldownSeconds = 0
	l := New(c)
	m := polymarket.Market{Slug: "match", ConditionID: "0xabc", End: now.Add(90 * time.Second), Active: true, UpToken: "up", DownToken: "down", Up: polymarket.Quote{Bid: .49, Ask: .50, BidDepth: 200, AskDepth: 200}}
	sig := strategy.Result{Ready: true, Direction: "UP"}
	l.Evaluate(m, sig, now)
	if len(l.Positions) != 1 || l.Positions[0].ConditionID != "0xabc" {
		t.Fatal("entry must record market identity")
	}
	// The market expires with no sell quote. Do not report fictional proceeds.
	m.Active = false
	l.Evaluate(m, strategy.Result{Direction: "WAIT"}, now.Add(91*time.Second))
	if l.Positions[0].Status != "PENDING_RESOLUTION" || l.Positions[0].MarkPnL != -l.Positions[0].EntryCost {
		t.Fatalf("bad unresolved accounting %+v", l.Positions[0])
	}
	bad := polymarket.Settlement{Slug: "match", ConditionID: "wrong", Winner: "UP"}
	if l.ApplySettlement(bad, now.Add(92*time.Second)) != 0 {
		t.Fatal("settled unrelated market")
	}
	result := polymarket.Settlement{Slug: "match", ConditionID: "0xabc", Winner: "DOWN"}
	if l.ApplySettlement(result, now.Add(92*time.Second)) != 1 || l.Positions[0].Status != "SETTLED" || l.Positions[0].RealizedPnL != -20 {
		t.Fatalf("bad loss settlement %+v", l)
	}
	cash := l.Cash
	if l.ApplySettlement(result, now.Add(94*time.Second)) != 0 || l.Cash != cash || l.TotalClosed != 1 {
		t.Fatal("duplicate settlement changed balances")
	}
	if l.OpenPositions() != 0 {
		t.Fatal("settled counted as open")
	}
}

func TestNoFeedStillExpiresInventory(t *testing.T) {
	now := time.Unix(1800000000, 0)
	l := New(DefaultConfig())
	l.Cash -= 20
	l.Positions = []Position{{Market: "m", Status: "OPEN", EntryCost: 20, MarketEnd: now.Add(-time.Second)}}
	if l.ExpireUnresolved(now) != 1 || l.Positions[0].Status != "PENDING_RESOLUTION" || l.Positions[0].MarkPnL != -20 {
		t.Fatal("outage must mark expired inventory zero")
	}
	if l.ExpireUnresolved(now) != 0 {
		t.Fatal("outage double processing")
	}
	l.RecordEquity(now)
	if l.CurrentEquity() != l.Config.StartingCash-20 {
		t.Fatalf("expired inventory must have zero equity; equity %v", l.CurrentEquity())
	}
}
