package maker

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
	"testing"
	"time"
)

func TestMakerQuotesAreConstrainedAndNonExecutable(t *testing.T) {
	now := time.Now()
	m := polymarket.Market{Slug: "m", Active: true, End: now.Add(3 * time.Minute), Up: polymarket.Quote{Bid: .47, Ask: .52, BidDepth: 100, AskDepth: 100, TickSize: .01}}
	p := Plan(m, strategy.Result{Ready: true, Direction: "UP"}, now, DefaultConfig())
	if p.Status != "INDICATIVE_ONLY_NO_FILLS" || p.Bid <= .47 || p.Ask >= .52 || p.Spread <= 0 {
		t.Fatalf("invalid plan: %+v", p)
	}
	m.Up.Ask = .471
	p = Plan(m, strategy.Result{Ready: true, Direction: "UP"}, now, DefaultConfig())
	if p.Status != "REJECTED" {
		t.Fatal("spread too tight")
	}
	m.Up.Ask = .52
	m.End = now.Add(10 * time.Second)
	if Plan(m, strategy.Result{Ready: true, Direction: "UP"}, now, DefaultConfig()).Status != "REJECTED" {
		t.Fatal("near expiry quote")
	}
}
