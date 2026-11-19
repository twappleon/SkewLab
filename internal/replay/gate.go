package replay

import "github.com/twappleon/SkewLab/internal/paper"

// EvidenceGate is a reproducible research readiness assessment, NEVER a
// permission to resume paper or real orders.
type EvidenceGate struct {
	Version           string        `json:"version"`
	Status            string        `json:"status"`
	TradingAuthorized bool          `json:"trading_authorized"`
	Ledger            LedgerAudit   `json:"ledger_audit"`
	Coverage          QuoteCoverage `json:"quote_coverage"`
	Pairs             []PairQuality `json:"paired_quality"`
	Blockers          []string      `json:"blockers"`
}

func BuildEvidenceGate(l *paper.Ledger, archive Archive, review Review) EvidenceGate {
	g := EvidenceGate{Version: "1.7.0", Status: "RESEARCH_BLOCKED", TradingAuthorized: false, Ledger: AuditLedger(l), Coverage: AssessQuoteCoverage(archive.Frames), Pairs: AssessPairs(review), Blockers: []string{}}
	if g.Ledger.Status != "RECORDED_ACCOUNTING_CONSISTENT_NOT_PROFIT_PROOF" {
		g.Blockers = append(g.Blockers, "LEDGER_ACCOUNTING_UNVERIFIED")
	}
	if len(archive.Frames) == 0 || g.Coverage.ValidFreshPairs == 0 {
		g.Blockers = append(g.Blockers, "NO_REPLAYABLE_FRESH_BOOKS")
	}
	if review.AllRecordedClosed != g.Ledger.RecordedClosed {
		g.Blockers = append(g.Blockers, "REPLAY_LEDGER_COUNT_MISMATCH")
	}
	if review.MissingOrUnverified > 0 {
		g.Blockers = append(g.Blockers, "UNVERIFIABLE_HISTORICAL_EXITS")
	}
	for _, p := range g.Pairs {
		if p.Window == "other" {
			continue
		}
		if p.Status != "EXPLORATORY_PAIRED_SAMPLE_NOT_VALIDATED" {
			g.Blockers = append(g.Blockers, "INSUFFICIENT_PAIRED_EVIDENCE_"+p.Window)
		}
	}
	if len(g.Blockers) == 0 {
		g.Status = "RESEARCH_SAMPLE_AVAILABLE_NOT_VALIDATED_OR_AUTHORIZED"
	}
	return g
}
