package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"testing"
	"time"
)

func TestAuditEntryNeverUsesFutureFrameQuote(t *testing.T) {
	n := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	cfg := paper.DefaultConfig()
	pos := paper.Position{EntryTime: n, EntryCost: 20, Qty: 40, EntryPrice: .502, Side: "UP", Market: "m", ConditionID: "c"}
	f := Frame{Time: n.Add(-2 * time.Second), Markets: []polymarket.Market{{Slug: "m", ConditionID: "c", Up: polymarket.Quote{Bid: .5, Ask: .5, BidDepth: 90, AskDepth: 90, UpdatedAt: n.Add(-time.Second)}}}}
	result := AuditEntry(pos, []Frame{f}, cfg)
	if result.RoundTripLossPct != nil || result.Status != "STALE_OR_INVALID_ENTRY_BOOK" {
		t.Fatalf("future quote disclosed early %+v", result)
	}
	f.Markets[0].Up.UpdatedAt = f.Time
	result = AuditEntry(pos, []Frame{f}, cfg)
	if result.Status != "OBSERVED_MODEL_ONLY" {
		t.Fatalf("valid quote lost %+v", result)
	}
}
