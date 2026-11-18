package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"testing"
	"time"
)

func TestStressScenariosDoNotClaimAnUnobservedEdge(t *testing.T) {
	r := StressCompare(nil, time.Now())
	if r.MinimumMatched != 0 || r.AllScenariosPositive || r.TradingAuthorized || len(r.Scenarios) != 3 {
		t.Fatal(r)
	}
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	xs := []experiment.Sample{}
	for i := 0; i < 115; i++ {
		at := base.Add(time.Duration(i) * 10 * time.Minute)
		s := fixture(at, 300, "UP")
		s.UpAsk = .53
		s.UpBid = .51
		s.DownAsk = .45
		s.DownBid = .43
		xs = append(xs, s)
	}
	r = StressCompare(xs, base.Add(1300*time.Minute))
	if r.MinimumMatched != 0 || r.AllScenariosPositive || r.Status != "INSUFFICIENT_MATCHED_EVIDENCE" {
		t.Fatal(r)
	}
	if r.Scenarios[2].Paired.ByWindow["5m"][0].TotalNetUSD > r.Scenarios[0].Paired.ByWindow["5m"][0].TotalNetUSD {
		t.Fatalf("higher costs improved hypothetical P&L %+v", r)
	}
}
