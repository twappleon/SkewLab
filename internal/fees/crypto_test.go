package fees

import (
	"math"
	"testing"
	"time"
)

func TestScheduleBoundaryAndFormula(t *testing.T) {
	old := ChangeUTC.Add(-time.Nanosecond)
	if CryptoCoefficient(old) != .07 || CryptoCoefficient(ChangeUTC) != .05 {
		t.Fatal("wrong schedule boundary")
	}
	if math.Abs(TakerFee(100, .5, old)-1.75) > 1e-9 {
		t.Fatal("old fee wrong")
	}
	if math.Abs(TakerFee(100, .5, ChangeUTC)-1.25) > 1e-9 {
		t.Fatal("new fee wrong")
	}
	if BufferedCost(100, .5, 0, old) != 51.75 {
		t.Fatal("cost wrong")
	}
	if BufferedProceeds(100, .5, 0, old) != 48.25 {
		t.Fatal("proceeds wrong")
	}
}
