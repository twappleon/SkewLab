package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"github.com/twappleon/SkewLab/internal/fees"
	"math"
	"sort"
	"time"
)

// FrictionWindow measures the loss of buying and immediately re-selling at the
// OBSERVED best prices, using hypothetical fee/slippage reserves. The source
// comparison archive does NOT record exit bid depth: these are NOT fills.
type FrictionWindow struct {
	Window                      string   `json:"window"`
	Observations                int      `json:"observations"`
	AlignedSignals              int      `json:"aligned_signals"`
	CostOnlyStopBreaches        int      `json:"cost_only_stop_breaches"`
	UnverifiedExitDepth         int      `json:"unverified_exit_depth"`
	MedianRoundTripLossPct      *float64 `json:"median_round_trip_loss_pct,omitempty"`
	MedianSettleBreakevenWinPct *float64 `json:"median_settlement_breakeven_win_pct,omitempty"`
}
type FrictionReport struct {
	Windows           []FrictionWindow `json:"windows"`
	StopTriggerPct    float64          `json:"stop_trigger_pct"`
	EntriesAuthorized bool             `json:"entries_authorized"`
	Limitation        string           `json:"limitation"`
}

func median(xs []float64) *float64 {
	if len(xs) == 0 {
		return nil
	}
	sort.Float64s(xs)
	m := xs[len(xs)/2]
	if len(xs)%2 == 0 {
		m = (xs[len(xs)/2-1] + m) / 2
	}
	v := math.Round(m*10000) / 10000
	return &v
}
func FrictionAudit(samples []experiment.Sample, now time.Time) FrictionReport {
	out := FrictionReport{StopTriggerPct: 5, EntriesAuthorized: false,
		Limitation: "No actual fills, queued orders or exit-bid depth in the observation archive. Same-quote liquidation is a modeled diagnostic; 5% is only a trigger, never a guaranteed stop."}
	out.Windows = []FrictionWindow{{Window: "5m"}, {Window: "15m"}}
	rt := [2][]float64{}
	hurdle := [2][]float64{}
	seen := map[string]bool{}
	for _, s := range samples {
		if s.Slug == "" || seen[s.Slug] || AuditIntegrity([]experiment.Sample{s}, now).Valid != 1 {
			continue
		}
		seen[s.Slug] = true
		i := 0
		if s.WindowSeconds == 900 {
			i = 1
		}
		w := &out.Windows[i]
		w.Observations++
		side := pick(s, 1)
		if side == "" {
			continue
		}
		w.AlignedSignals++
		ask, bid := s.UpAsk, s.UpBid
		if side == "DOWN" {
			ask, bid = s.DownAsk, s.DownBid
		}
		entry := ask + .002
		exit := bid - .002
		if !(entry > 0 && entry < 1 && exit > 0 && exit < 1) {
			continue
		}
		cost := fees.BufferedCost(1, entry, 25, s.CapturedAt)
		proceeds := fees.BufferedProceeds(1, exit, 25, s.CapturedAt)
		if !finite(cost) || !finite(proceeds) || cost <= 0 {
			continue
		}
		friction := 100 * (cost - proceeds) / cost
		rt[i] = append(rt[i], friction)
		hurdle[i] = append(hurdle[i], 100*cost)
		if friction >= out.StopTriggerPct {
			w.CostOnlyStopBreaches++
		}
		// An ask-depth observation never proves a later exit would have filled.
		w.UnverifiedExitDepth++
	}
	for i := range out.Windows {
		out.Windows[i].MedianRoundTripLossPct = median(rt[i])
		out.Windows[i].MedianSettleBreakevenWinPct = median(hurdle[i])
	}
	return out
}
