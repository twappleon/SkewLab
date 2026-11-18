package strategy

import (
	"testing"
	"time"
)

func TestAlignedUpAndDown(t *testing.T) {
	now := time.Unix(1800000000, 0)
	for _, tt := range []struct {
		delta float64
		want  string
	}{{3, "UP"}, {-3, "DOWN"}} {
		ticks := make([]Tick, 0, 61)
		for i := 0; i <= 60; i++ {
			ticks = append(ticks, Tick{Time: now.Add(time.Duration(i-60) * time.Second), Price: 100000 + float64(i)*tt.delta, Source: "test"})
		}
		got := Calculate(ticks, now, 0.02, 0.2)
		if got.Direction != tt.want || !got.Ready {
			t.Fatalf("wanted %s got %+v", tt.want, got)
		}
	}
}
func TestNoSignalWithoutAnchorOrOnStaleFeed(t *testing.T) {
	now := time.Unix(1800000000, 0)
	ticks := []Tick{}
	for i := 0; i < 21; i++ {
		ticks = append(ticks, Tick{Time: now.Add(time.Duration(i-20) * time.Second), Price: 100000 + float64(i)*10, Source: "test"})
	}
	if r := Calculate(ticks, now, 0.02, 0.2); r.Ready {
		t.Fatalf("incorrectly ready: %+v", r)
	}
	ticks = nil
	for i := 0; i < 61; i++ {
		ticks = append(ticks, Tick{Time: now.Add(time.Duration(i-70) * time.Second), Price: 100000 + float64(i)*10, Source: "test"})
	}
	if r := Calculate(ticks, now, 0.02, 0.2); r.Ready {
		t.Fatalf("accepted stale: %+v", r)
	}
}
func TestNoVenueStitching(t *testing.T) {
	now := time.Unix(1800000000, 0)
	ticks := []Tick{}
	for i := 0; i < 40; i++ {
		src := "Binance"
		if i >= 30 {
			src = "Coinbase"
		}
		ticks = append(ticks, Tick{Time: now.Add(time.Duration(i-39) * time.Second), Price: 100000 + float64(i)*10, Source: src})
	}
	if r := Calculate(ticks, now, 0.02, 0.2); r.Ready {
		t.Fatalf("combined exchange prices: %+v", r)
	}
}
