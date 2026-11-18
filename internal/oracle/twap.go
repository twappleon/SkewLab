// Package oracle independently audits captured Chainlink TWAP-60s ticks.
// The package never substitutes exchange spot price for Chainlink and its
// result is not a substitute for the official Polymarket resolved outcome.
package oracle

import (
	"errors"
	"fmt"
	"math"
	"time"
)

type Tick struct {
	Time   time.Time `json:"time"`
	Price  float64   `json:"price"`
	Source string    `json:"source"`
}

type Audit struct {
	WindowStart time.Time `json:"window_start"`
	WindowEnd   time.Time `json:"window_end"`
	OpenTWAP    float64   `json:"open_twap"`
	CloseTWAP   float64   `json:"close_twap"`
	Direction   string    `json:"direction"`
	Comparable  bool      `json:"comparable"`
}

// Compare requires the source name and exact UTC window-edge timestamps.
// Missing, duplicated-conflicting or invalid ticks fail closed rather than
// using a nearby spot sample or a post-close tick (no lookahead).
func Compare(ticks []Tick, start time.Time, window time.Duration) (Audit, error) {
	res := Audit{WindowStart: start.UTC(), WindowEnd: start.UTC().Add(window)}
	if window != 5*time.Minute && window != 15*time.Minute {
		return res, errors.New("only 5m/15m TWAP markets supported")
	}
	if start.IsZero() || start.Unix()%int64(window/time.Second) != 0 {
		return res, errors.New("misaligned market window")
	}
	foundStart, foundEnd := false, false
	for _, t := range ticks {
		if t.Source != "chainlink-twap-60s" {
			continue
		}
		if t.Price <= 0 || math.IsInf(t.Price, 0) || math.IsNaN(t.Price) {
			continue
		}
		at := t.Time.UTC()
		if at.Equal(res.WindowStart) {
			if foundStart && t.Price != res.OpenTWAP {
				return res, errors.New("conflicting start TWAP")
			}
			res.OpenTWAP = t.Price
			foundStart = true
		}
		if at.Equal(res.WindowEnd) {
			if foundEnd && t.Price != res.CloseTWAP {
				return res, errors.New("conflicting end TWAP")
			}
			res.CloseTWAP = t.Price
			foundEnd = true
		}
	}
	if !foundStart || !foundEnd {
		return res, fmt.Errorf("missing exact Chainlink TWAP market boundary: start=%t end=%t", foundStart, foundEnd)
	}
	res.Direction = "DOWN"
	if res.CloseTWAP >= res.OpenTWAP {
		res.Direction = "UP"
	}
	res.Comparable = true
	return res, nil
}
