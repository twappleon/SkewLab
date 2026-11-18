package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"testing"
	"time"
)

func TestPairedDoesNotLookAheadAndSharesTheSameSnapshot(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	samples := []experiment.Sample{}
	for i := 0; i < 120; i++ {
		at := base.Add(time.Duration(i) * 10 * time.Minute)
		s := fixture(at, 300, "UP")
		s.Slug = at.Format(time.RFC3339Nano)
		s.DownBid = .43
		s.DownAsk = .45
		s.UpAsk = .53
		s.UpBid = .51
		s.ResolvedAt = s.MarketEnd.Add(time.Second)
		samples = append(samples, s)
	}
	r := PairedCheck(samples, base.Add(1300*time.Minute), .002, 25)
	x := r.ByWindow["5m"][0]
	if x.MatchedSettled != 20 || x.IndependentUTCDays != 1 || x.ConservativeDayBlockLowerUSD != nil || r.TradingAuthorized {
		t.Fatalf("misrepresented forward research: %+v", x)
	}
	for i := range samples {
		samples[i].ResolvedAt = base.Add(1400 * time.Minute)
	}
	r = PairedCheck(samples, base.Add(1500*time.Minute), .002, 25)
	if r.ByWindow["5m"][0].MatchedSettled != 0 {
		t.Fatal("used future result to qualify a past decision")
	}
}
