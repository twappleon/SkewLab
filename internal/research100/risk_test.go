package research100

import (
	"testing"
)

func TestRiskUnderstandsBinaryTotalLoss(t *testing.T) {
	a := AssessWorstCaseRisk(100, 5, 0, 0, 2)
	if a.Status != "WORST_CASE_EXCEEDS_DAILY_STOP" || a.TradingAuthorized {
		t.Fatalf("should reject $5 loss with $2 cap: %+v", a)
	}
	b := AssessWorstCaseRisk(100, 1.5, 0, .25, 2)
	if b.Status != "WITHIN_RESEARCH_BUDGET_NOT_TRADE_AUTHORIZATION" || b.TradingAuthorized {
		t.Fatalf("unexpected risk gate %+v", b)
	}
	c := AssessWorstCaseRisk(100, 1.5, 1, 0, 2)
	if c.Status != "WORST_CASE_EXCEEDS_DAILY_STOP" {
		t.Fatalf("open exposure ignored %+v", c)
	}
}
