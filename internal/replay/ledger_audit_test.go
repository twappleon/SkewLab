package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"testing"
	"time"
)

func TestAuditLedgerDetectsRewritingOrMissingLosses(t *testing.T) {
	n := time.Date(2026, 11, 19, 0, 0, 0, 0, time.UTC)
	l := paper.New(paper.DefaultConfig())
	l.Cash = 970
	l.RealizedPnL = -30
	l.TotalClosed = 1
	l.Positions = []paper.Position{{Status: "CLOSED", EntryCost: 20, RealizedPnL: -30, ExitTime: &n}}
	a := AuditLedger(l)
	if a.Status != "RECORDED_ACCOUNTING_CONSISTENT_NOT_PROFIT_PROOF" || a.TradingAuthorized {
		t.Fatalf("valid historical loss rejected: %+v", a)
	}
	l.Cash = 1000
	if a = AuditLedger(l); a.Status != "INCOMPLETE_OR_INVALID_LEDGER" {
		t.Fatal("cash rewrite accepted")
	}
	l.Cash = 970
	l.TotalClosed = 2
	if a = AuditLedger(l); a.Status != "INCOMPLETE_OR_INVALID_LEDGER" {
		t.Fatal("missing record accepted")
	}
}
