package research100

import (
	"math"
	"testing"
	"time"
)

func TestCostToResolutionHonorsFeeScheduleAndNoFreeLunch(t *testing.T) {
	before := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	after := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
	a := CostToResolution(.5, .002, 25, 2, before)
	b := CostToResolution(.5, .002, 25, 2, after)
	if a.Status != "COST_MODEL_ONLY_NOT_EXECUTABLE" || b.Status != "COST_MODEL_ONLY_NOT_EXECUTABLE" || !(a.BreakEvenWinRatePct > b.BreakEvenWinRatePct) {
		t.Fatalf("fee calendar ignored: %+v vs %+v", a, b)
	}
	if a.MinimumWinRateWithBufferPct <= a.BreakEvenWinRatePct {
		t.Fatal("missing safety margin")
	}
	if CostToResolution(math.NaN(), 0, 25, 0, before).Status != "INVALID_COST_INPUT" {
		t.Fatal("NaN accepted")
	}
}
