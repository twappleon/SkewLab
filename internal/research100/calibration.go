package research100

import (
	"math"
	"time"
)

// Calibration describes a quote-as-probability proxy, NOT calibrated truth.
// Only official outcomes known by the as-of timestamp are scored.
type Calibration struct {
	Samples  int      `json:"samples"`
	Brier    *float64 `json:"brier,omitempty"`
	LogLoss  *float64 `json:"log_loss,omitempty"`
	BaseRate *float64 `json:"base_rate,omitempty"`
	Status   string   `json:"status"`
}

func ScoreCalibration(d Dataset, asOf time.Time) Calibration {
	out := Calibration{Status: "NO_RESOLVED_SAMPLES"}
	var sq, loss float64
	wins := 0
	for _, o := range d.Items {
		if o.Winner == "" || o.ResolvedAt.After(asOf) || o.CapturedAt.After(asOf) {
			continue
		}
		p := (o.Ask + o.Bid) / 2 // midpoint proxy, not a proof of true probability
		if p <= 0 || p >= 1 || !finite(p) {
			continue
		}
		y := 0.0
		if o.Winner == o.Side {
			y = 1
			wins++
		}
		delta := p - y
		sq += delta * delta
		loss -= y*math.Log(p) + (1-y)*math.Log(1-p)
		out.Samples++
	}
	if out.Samples == 0 {
		return out
	}
	n := float64(out.Samples)
	b, l, r := sq/n, loss/n, float64(wins)/n
	out.Brier = &b
	out.LogLoss = &l
	out.BaseRate = &r
	out.Status = "QUOTE_PROXY_RESEARCH_ONLY"
	return out
}
