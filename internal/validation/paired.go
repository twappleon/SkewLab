package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"github.com/twappleon/SkewLab/internal/fees"
	"math"
	"sort"
	"time"
)

type PairedRule struct {
	Rule                         string   `json:"rule"`
	ProspectiveObservations      int      `json:"prospective_observations"`
	MatchedSettled               int      `json:"matched_settled"`
	Correct                      int      `json:"correct"`
	TotalNetUSD                  float64  `json:"total_net_per_share_usd"`
	MatchedBaselineNetUSD        float64  `json:"matched_baseline_net_per_share_usd"`
	MeanExcessUSD                *float64 `json:"mean_excess_usd,omitempty"`
	ConservativeDayBlockLowerUSD *float64 `json:"conservative_day_block_lower_usd,omitempty"`
	IndependentUTCDays           int      `json:"observed_utc_days"`
	Status                       string   `json:"status"`
}
type PairedReport struct {
	ByWindow          map[string][]PairedRule `json:"by_window"`
	Status            string                  `json:"status"`
	TradingAuthorized bool                    `json:"trading_authorized"`
}

func pick(s experiment.Sample, k int) string {
	if k == 2 {
		up := (s.UpBid + s.UpAsk) / 2
		down := (s.DownBid + s.DownAsk) / 2
		if math.Abs(up-down) < 1e-9 {
			return ""
		}
		if up > down {
			return "UP"
		}
		return "DOWN"
	}
	if s.MomentumPct >= .02 && (k == 0 || s.Skew >= .20) {
		return "UP"
	}
	if s.MomentumPct <= -.02 && (k == 0 || s.Skew <= -.20) {
		return "DOWN"
	}
	return ""
}
func buyCost(s experiment.Sample, side string, slippage, bps float64) float64 {
	ask := s.UpAsk
	if side == "DOWN" {
		ask = s.DownAsk
	}
	if ask <= 0 || ask+slippage >= 1 {
		return math.NaN()
	}
	return fees.BufferedCost(1, ask+slippage, bps, s.CapturedAt)
}
func net(s experiment.Sample, side string, slippage, bps float64) float64 {
	cost := buyCost(s, side, slippage, bps)
	if !finite(cost) {
		return math.NaN()
	}
	if s.Winner == side {
		return 1 - cost
	}
	return -cost
}

type ckey struct {
	window int
	side   string
	bucket int
}

func bidBucket(ask float64) int {
	v := int(math.Floor(ask * 5))
	if v < 0 {
		return 0
	}
	if v > 4 {
		return 4
	}
	return v
}
func priceOf(s experiment.Sample, side string) float64 {
	if side == "DOWN" {
		return s.DownAsk
	}
	return s.UpAsk
}
func PairedCheck(samples []experiment.Sample, now time.Time, slip, bps float64) PairedReport {
	out := PairedReport{Status: "INSUFFICIENT_PROSPECTIVE_EVIDENCE", ByWindow: map[string][]PairedRule{}, TradingAuthorized: false}
	if slip < 0 || bps < 0 || !finite(slip) || !finite(bps) {
		out.Status = "INVALID_COST_ASSUMPTIONS"
		return out
	}
	for _, w := range []string{"5m", "15m"} {
		out.ByWindow[w] = []PairedRule{{Rule: "momentum_only", Status: "INSUFFICIENT"}, {Rule: "momentum_semivariance", Status: "INSUFFICIENT"}}
	}
	ordered := []experiment.Sample{}
	seen := map[string]bool{}
	resolved := []experiment.Sample{}
	for _, s := range samples {
		if s.Slug == "" || seen[s.Slug] || AuditIntegrity([]experiment.Sample{s}, now).Valid != 1 {
			continue
		}
		seen[s.Slug] = true
		ordered = append(ordered, s)
		if s.Winner != "" && !s.ResolvedAt.IsZero() {
			resolved = append(resolved, s)
		}
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CapturedAt.Before(ordered[j].CapturedAt) })
	sort.Slice(resolved, func(i, j int) bool { return resolved[i].ResolvedAt.Before(resolved[j].ResolvedAt) })
	prior := map[ckey]int{}
	index := 0
	daily := map[string][]map[string][]float64{"5m": {{}, {}}, "15m": {{}, {}}}
	for _, s := range ordered {
		for index < len(resolved) && resolved[index].ResolvedAt.Before(s.CapturedAt) {
			h := resolved[index]
			index++
			for _, side := range []string{"UP", "DOWN"} {
				prior[ckey{h.WindowSeconds, side, bidBucket(priceOf(h, side))}]++
			}
		}
		w := "5m"
		if s.WindowSeconds == 900 {
			w = "15m"
		}
		base := pick(s, 2)
		if base == "" {
			continue
		}
		for k := 0; k < 2; k++ {
			side := pick(s, k)
			if side == "" {
				continue
			}
			if prior[ckey{s.WindowSeconds, side, bidBucket(priceOf(s, side))}] < 100 {
				continue
			}
			rule := &out.ByWindow[w][k]
			rule.ProspectiveObservations++
			if s.Winner == "" {
				continue
			}
			v, b := net(s, side, slip, bps), net(s, base, slip, bps)
			if !finite(v) || !finite(b) {
				continue
			}
			rule.MatchedSettled++
			if side == s.Winner {
				rule.Correct++
			}
			rule.TotalNetUSD += v
			rule.MatchedBaselineNetUSD += b
			day := s.CapturedAt.UTC().Format("2006-01-02")
			daily[w][k][day] = append(daily[w][k][day], v-b)
		}
	}
	for w, rs := range out.ByWindow {
		for k := range rs {
			r := &rs[k]
			if r.MatchedSettled == 0 {
				continue
			}
			excess := r.TotalNetUSD - r.MatchedBaselineNetUSD
			mean := math.Round(10000*excess/float64(r.MatchedSettled)) / 10000
			r.MeanExcessUSD = &mean
			r.TotalNetUSD = math.Round(r.TotalNetUSD*10000) / 10000
			r.MatchedBaselineNetUSD = math.Round(r.MatchedBaselineNetUSD*10000) / 10000
			r.IndependentUTCDays = len(daily[w][k])
			if r.MatchedSettled >= 30 && r.IndependentUTCDays >= 10 {
				blocks := make([]float64, 0, len(daily[w][k]))
				for _, d := range daily[w][k] {
					total := 0.0
					for _, v := range d {
						total += v
					}
					blocks = append(blocks, total/float64(len(d)))
				}
				mu := 0.0
				for _, v := range blocks {
					mu += v
				}
				mu /= float64(len(blocks))
				variance := 0.0
				for _, v := range blocks {
					variance += (v - mu) * (v - mu)
				}
				variance /= float64(len(blocks) - 1)
				lower := math.Round(10000*(mu-2.58*math.Sqrt(variance/float64(len(blocks))))) / 10000
				r.ConservativeDayBlockLowerUSD = &lower
				r.Status = "DAY_BLOCK_EXPLORATORY_ONLY"
			} else {
				r.Status = "WAITING_FOR_30_MATCHED_AND_10_DAYS"
			}
		}
		out.ByWindow[w] = rs
	}
	return out
}
