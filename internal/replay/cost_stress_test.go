package replay

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"testing"
	"time"
)

func TestCostStressRejectsCostOnlyChurn(t *testing.T) {
	n := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	q := polymarket.Quote{Bid: .49, Ask: .51, BidDepth: 200, AskDepth: 200, UpdatedAt: n}
	a := EvaluateCostStress(q, n, 40, .002, 25, 5)
	if a.Status != "RESEARCH_ONLY" || !a.StopAlreadyConsumed || *a.NormalLossPct < 5 || *a.DoubledSlippageLossPct < *a.NormalLossPct || *a.QuadrupledReserveLossPct < *a.NormalLossPct {
		t.Fatalf("cost stress wrong %+v", a)
	}
	q.BidDepth = 1
	if b := EvaluateCostStress(q, n, 40, .002, 25, 5); b.NormalLossPct != nil {
		t.Fatalf("extrapolated insufficient depth: %+v", b)
	}
}
