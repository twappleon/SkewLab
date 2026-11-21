package replay

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"testing"
	"time"
)

func TestCoverageNoUnobservedInference(t *testing.T) {
	n := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	q := polymarket.Quote{Bid: .4, Ask: .41, BidDepth: 10, AskDepth: 10, UpdatedAt: n}
	mk := func(slug string, active bool, down polymarket.Quote) Frame {
		return Frame{Time: n, Markets: []polymarket.Market{{Slug: slug, Active: active, Up: q, Down: down}}}
	}
	c := AssessQuoteCoverage([]Frame{{Time: n}, mk("btc-updown-5m-123", true, q), mk("btc-updown-15m-123", true, polymarket.Quote{})})
	if c.RecordedFrames != 3 || c.MissingPairs != 1 || c.ValidFreshPairs != 1 || c.InvalidOrStalePairs != 1 || c.TradeEvidenceReady {
		t.Fatalf("wrong coverage %+v", c)
	}
	if empty := AssessQuoteCoverage(nil); empty.ValidFreshPairs != 0 || empty.TradeEvidenceReady {
		t.Fatal("invented coverage")
	}
}
