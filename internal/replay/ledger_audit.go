package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"math"
)

// LedgerAudit compares redundant recorded totals. It never modifies balances,
// recomputes missing fills, or treats accounting agreement as actual profit.
type LedgerAudit struct {
	Status                   string  `json:"status"`
	RecordedClosed           int     `json:"recorded_closed"`
	LedgerClosed             int     `json:"ledger_closed"`
	RecomputedRealizedPnLUSD float64 `json:"recomputed_realized_pnl_usd"`
	ReportedRealizedPnLUSD   float64 `json:"reported_realized_pnl_usd"`
	CashResidualUSD          float64 `json:"cash_residual_usd"`
	TradingAuthorized        bool    `json:"trading_authorized"`
}

func AuditLedger(l *paper.Ledger) LedgerAudit {
	out := LedgerAudit{Status: "INCOMPLETE_OR_INVALID_LEDGER", TradingAuthorized: false}
	if l == nil || !finite(l.Config.StartingCash) || l.Config.StartingCash <= 0 || !finite(l.Cash) || l.Cash < 0 || !finite(l.RealizedPnL) {
		return out
	}
	out.LedgerClosed = l.TotalClosed
	out.ReportedRealizedPnLUSD = l.RealizedPnL
	realized := 0.0
	openCost := 0.0
	for _, p := range l.Positions {
		switch p.Status {
		case "CLOSED", "SETTLED":
			if p.ExitTime == nil || !finite(p.RealizedPnL) {
				return out
			}
			out.RecordedClosed++
			realized += p.RealizedPnL
		case "OPEN", "PENDING_RESOLUTION":
			if !finite(p.EntryCost) || p.EntryCost <= 0 {
				return out
			}
			openCost += p.EntryCost
		default:
			return out
		}
	}
	out.RecomputedRealizedPnLUSD = math.Round(realized*10000) / 10000
	out.CashResidualUSD = math.Round((l.Cash-(l.Config.StartingCash+realized-openCost))*10000) / 10000
	if out.RecordedClosed != out.LedgerClosed || math.Abs(out.RecomputedRealizedPnLUSD-out.ReportedRealizedPnLUSD) > .02 || math.Abs(out.CashResidualUSD) > .02 {
		return out
	}
	out.Status = "RECORDED_ACCOUNTING_CONSISTENT_NOT_PROFIT_PROOF"
	return out
}
