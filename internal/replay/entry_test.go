package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"math"
	"testing"
	"time"
)

func TestEntryFrictionAuditedNotFabricated(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	cfg := paper.DefaultConfig()
	p := paper.Position{Market: "btc-updown-5m-111", Side: "UP", EntryTime: now, EntryPrice: .502, EntryCost: 20, Qty: 20 / .52, ConditionID: "id"}
	m := polymarket.Market{Slug: p.Market, ConditionID: "id", Up: polymarket.Quote{Bid: .50, Ask: .50, BidDepth: 100, AskDepth: 100, UpdatedAt: now.Add(-time.Second)}}
	f := Frame{Time: now.Add(-time.Second), Markets: []polymarket.Market{m}}
	a := AuditEntry(p, []Frame{f}, cfg)
	if a.Status != "OBSERVED_MODEL_ONLY" || a.RoundTripLossPct == nil || !*&a.CostOnlyStopBreach || *a.RoundTripLossPct <= 5 {
		t.Fatalf("bad roundtrip: %+v", a)
	}
	if b := AuditEntry(p, []Frame{{Time: now.Add(time.Second), Markets: []polymarket.Market{m}}}, cfg); b.RoundTripLossPct != nil {
		t.Fatal("future quote used")
	}
	m.Up.UpdatedAt = now.Add(-time.Minute)
	if b := AuditEntry(p, []Frame{{Time: now, Markets: []polymarket.Market{m}}}, cfg); b.RoundTripLossPct != nil {
		t.Fatal("stale quote used")
	}
	p.EntryCost = math.NaN()
	if b := AuditEntry(p, []Frame{f}, cfg); b.Status != "INVALID_POSITION_OR_CONFIG" {
		t.Fatal("NaN accepted")
	}
}
