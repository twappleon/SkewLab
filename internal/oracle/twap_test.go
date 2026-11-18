package oracle

import (
	"testing"
	"time"
)

func TestCompareRequiresOnlyExactTWAPBoundaries(t *testing.T) {
	start := time.Unix(1800000000, 0)
	ticks := []Tick{{Time: start, Price: 100, Source: "chainlink-twap-60s"}, {Time: start.Add(5 * time.Minute), Price: 100, Source: "chainlink-twap-60s"}}
	result, err := Compare(ticks, start, 5*time.Minute)
	if err != nil || result.Direction != "UP" {
		t.Fatalf("equal TWAP must be UP %+v %v", result, err)
	}
	ticks[1].Price = 99
	result, err = Compare(ticks, start, 5*time.Minute)
	if err != nil || result.Direction != "DOWN" {
		t.Fatalf("fall must be DOWN %+v %v", result, err)
	}
	ticks[0].Source = "binance-spot"
	if _, err = Compare(ticks, start, 5*time.Minute); err == nil {
		t.Fatal("spot substituted for Chainlink")
	}
	ticks[0].Source = "chainlink-twap-60s"
	ticks[1].Time = ticks[1].Time.Add(time.Second)
	if _, err = Compare(ticks, start, 5*time.Minute); err == nil {
		t.Fatal("non-exact boundary accepted")
	}
}
