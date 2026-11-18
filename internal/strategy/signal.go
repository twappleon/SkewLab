package strategy

import (
	"math"
	"time"
)

type Tick struct {
	Time   time.Time `json:"time"`
	Price  float64   `json:"price"`
	Source string    `json:"source"`
}

type Result struct {
	Direction        string  `json:"direction"`
	MomentumPct      float64 `json:"momentum_pct"`
	Skew             float64 `json:"skew"`
	UpSemivariance   float64 `json:"up_semivariance"`
	DownSemivariance float64 `json:"down_semivariance"`
	Samples          int     `json:"samples"`
	Ready            bool    `json:"ready"`
	Reason           string  `json:"reason"`
}

// Calculate requires a real 30-second price anchor and computes realized
// upside/downside semivariance over the trailing 60 seconds. No signal is
// emitted when the input is stale, sparse, or switches venues.
func Calculate(ticks []Tick, now time.Time, minMomentumPct, minSkew float64) Result {
	out := Result{Direction: "WAIT", Reason: "warming up"}
	if len(ticks) < 12 {
		return out
	}
	valid := make([]Tick, 0, len(ticks))
	source := ticks[len(ticks)-1].Source
	for _, t := range ticks {
		if t.Price > 0 && t.Source == source && !t.Time.After(now) && now.Sub(t.Time) <= 60*time.Second {
			valid = append(valid, t)
		}
	}
	out.Samples = len(valid)
	if len(valid) < 12 {
		out.Reason = "insufficient samples"
		return out
	}
	newest := valid[len(valid)-1]
	if now.Sub(newest.Time) > 5*time.Second {
		out.Reason = "stale BTC feed"
		return out
	}
	var anchor *Tick
	for i := range valid {
		if !valid[i].Time.After(newest.Time.Add(-30 * time.Second)) {
			anchor = &valid[i]
		}
	}
	if anchor == nil || newest.Time.Sub(anchor.Time) > 35*time.Second {
		out.Reason = "missing 30s anchor"
		return out
	}
	var up, down float64
	validReturns := 0
	for i := 1; i < len(valid); i++ {
		dt := valid[i].Time.Sub(valid[i-1].Time)
		if dt <= 0 || dt > 8*time.Second {
			out.Reason = "BTC feed gap"
			return out
		}
		r := math.Log(valid[i].Price / valid[i-1].Price)
		if r > 0 {
			up += r * r
		} else {
			down += r * r
		}
		validReturns++
	}
	if validReturns < 11 || up+down < 1e-16 {
		out.Reason = "no meaningful volatility"
		return out
	}
	out.MomentumPct = (newest.Price/anchor.Price - 1) * 100
	out.UpSemivariance = up
	out.DownSemivariance = down
	out.Skew = (up - down) / (up + down)
	out.Ready = true
	out.Reason = "direction and semivariance disagree"
	if out.MomentumPct >= minMomentumPct && out.Skew >= minSkew {
		out.Direction = "UP"
		out.Reason = "momentum and upside semivariance aligned"
	} else if out.MomentumPct <= -minMomentumPct && out.Skew <= -minSkew {
		out.Direction = "DOWN"
		out.Reason = "momentum and downside semivariance aligned"
	}
	return out
}
