package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"testing"
	"time"
)

func TestFrictionShowsCostOnlyStopEvenWithoutPriceMove(t *testing.T) {
	now := time.Date(2026, 11, 18, 11, 0, 0, 0, time.UTC)
	s := fixture(now.Add(-time.Hour), 300, "UP")
	s.UpBid, s.UpAsk = .499, .50
	s.DownBid, s.DownAsk = .499, .50
	x := FrictionAudit([]experiment.Sample{s}, now)
	w := x.Windows[0]
	if w.AlignedSignals != 1 || w.CostOnlyStopBreaches != 1 || w.UnverifiedExitDepth != 1 || w.MedianRoundTripLossPct == nil || *w.MedianRoundTripLossPct <= 5 || w.MedianSettleBreakevenWinPct == nil || *w.MedianSettleBreakevenWinPct <= 50 || x.EntriesAuthorized {
		t.Fatalf("cost error: %+v", x)
	}
}
func TestFrictionDoesNotInventDataOrCreditFuture(t *testing.T) {
	now := time.Now().UTC()
	x := FrictionAudit(nil, now)
	if x.Windows[0].MedianRoundTripLossPct != nil || x.Windows[0].CostOnlyStopBreaches != 0 {
		t.Fatal(x)
	}
	s := fixture(now.Add(time.Minute), 300, "UP")
	x = FrictionAudit([]experiment.Sample{s}, now)
	if x.Windows[0].Observations != 0 {
		t.Fatal("future sample leaked")
	}
}
