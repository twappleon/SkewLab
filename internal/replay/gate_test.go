package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"testing"
)

func TestEvidenceGateFailsClosedWhenNoActualMarketEvidence(t *testing.T) {
	l := paper.New(paper.DefaultConfig())
	r := BuildReview(l.Positions, l.Config, Archive{})
	g := BuildEvidenceGate(l, Archive{}, r)
	if g.TradingAuthorized || g.Status != "RESEARCH_BLOCKED" || len(g.Blockers) < 2 {
		t.Fatalf("unsafe empty-data gate: %+v", g)
	}
}
func TestEvidenceGateCannotIssueTradingPermission(t *testing.T) {
	l := paper.New(paper.DefaultConfig())
	r := Review{AllRecordedClosed: 0, Groups: []Group{{Window: "5m", PairedModeledExits: 100, IndependentUTCDays: 20}, {Window: "15m", PairedModeledExits: 100, IndependentUTCDays: 20}}}
	g := BuildEvidenceGate(l, Archive{}, r)
	if g.TradingAuthorized || g.Status != "RESEARCH_BLOCKED" {
		t.Fatal("spurious permission")
	}
}
