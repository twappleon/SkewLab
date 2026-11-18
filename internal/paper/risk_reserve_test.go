package paper

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
	"math"
	"testing"
	"time"
)

func TestWorstCaseFullStakeNotFivePercentStop(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	c := DefaultConfig()
	c.StopLossPct = 15
	c.CooldownSeconds = 0
	l := New(c)
	mk := func(slug string) polymarket.Market {
		return polymarket.Market{Slug: slug, End: now.Add(4 * time.Minute), Active: true, UpToken: "UP", Up: polymarket.Quote{Bid: .49, Ask: .50, BidDepth: 1000, AskDepth: 1000}}
	}
	if got := l.Evaluate(mk("m1"), strategy.Result{Direction: "UP", Ready: true}, now); len(got) != 1 {
		t.Fatalf("first entry blocked: %v", got)
	}
	if l.CanReserveNewStake(now) || l.WorstCaseCommittedLossUSD(now) != 20 {
		t.Fatalf("did not reserve full $20 stake: %v", l.WorstCaseCommittedLossUSD(now))
	}
	l.Evaluate(mk("m2"), strategy.Result{Direction: "UP", Ready: true}, now.Add(time.Second))
	if len(l.Positions) != 1 {
		t.Fatal("second entry exceeded $30 potential-loss budget")
	}
}
func TestPriorLossesConsumeBudgetButPriorWinsDoNotReplenish(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	l := New(DefaultConfig())
	l.Positions = []Position{{Status: "SETTLED", ExitTime: &now, RealizedPnL: -20}, {Status: "CLOSED", ExitTime: &now, RealizedPnL: 100}}
	if l.CanReserveNewStake(now) {
		t.Fatal("must not net profits against a $20 realized loss")
	}
	tomorrow := now.Add(24 * time.Hour)
	if !l.CanReserveNewStake(tomorrow) {
		t.Fatal("new UTC day should clear historical realized-loss budget")
	}
	l.Positions = append(l.Positions, Position{Status: "PENDING_RESOLUTION", EntryCost: math.NaN()})
	if l.CanReserveNewStake(tomorrow) {
		t.Fatal("corrupt open stake must fail closed")
	}
}
