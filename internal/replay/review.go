package replay

import (
	"github.com/twappleon/SkewLab/internal/paper"
	"math"
	"strings"
)

type Group struct {
	Window                     string  `json:"window"`
	RecordedClosedTrades       int     `json:"recorded_closed_trades"`
	VerifiedEntrySnapshots     int     `json:"entry_snapshots"`
	MissingOrUnverifiedEntries int     `json:"missing_or_unverified_entries"`
	CostOnlyStopRisk           int     `json:"cost_only_stop_risk"`
	PairedModeledExits         int     `json:"paired_modeled_exits"`
	IndependentUTCDays         int     `json:"independent_utc_days"`
	RecordedPnLOnPairsUSD      float64 `json:"recorded_pnl_on_pairs_usd"`
	LegacyModeledPnLUSD        float64 `json:"legacy_modeled_pnl_usd"`
	AlternativeModeledPnLUSD   float64 `json:"alternative_modeled_pnl_usd"`
}
type Review struct {
	Evidence            *EvidenceGate `json:"evidence_gate,omitempty"`
	ModelVersion        string        `json:"model_version"`
	StudyOnly           bool          `json:"study_only"`
	TradingAuthorized   bool          `json:"trading_authorized"`
	Status              string        `json:"status"`
	ArchiveFrames       int           `json:"archive_frames"`
	ArchivedFiles       int           `json:"archived_files"`
	Groups              []Group       `json:"groups"`
	MatchedPairs        int           `json:"matched_pairs"`
	AllRecordedClosed   int           `json:"all_recorded_closed"`
	MissingOrUnverified int           `json:"missing_or_unverified"`
	Limitation          string        `json:"limitation"`
}

func windowOf(market string) string {
	if strings.HasPrefix(market, "btc-updown-5m-") {
		return "5m"
	}
	if strings.HasPrefix(market, "btc-updown-15m-") {
		return "15m"
	}
	return "other"
}
func cleanMoney(v float64) float64 { return math.Round(v*10000) / 10000 }

// BuildReview compares two *predeclared* exit assumptions on identical
// historical observations for the very same recorded entry. Zero matched pairs
// is explicitly NOT zero performance. Old exchange-filling history cannot be
// reconstructed after the fact and its recorded P&L is never overwritten.
func BuildReview(positions []paper.Position, cfg paper.Config, archive Archive) Review {
	r := Review{ModelVersion: "1.6.0", StudyOnly: true, TradingAuthorized: false, Status: "INSUFFICIENT_INDEPENDENT_REPLAY_EVIDENCE", ArchiveFrames: len(archive.Frames), ArchivedFiles: archive.Files,
		Limitation: "All fills are hypothetical at recorded best bid/ask with modeled fees/slippage and displayed depth; no guarantee of executable quantity. Missing/rotated pre-entry or exit snapshots are unverified, not losses or gains. The alternative stop does NOT authorize a trading change, and replayed exits do not prove observed historical execution."}
	r.Groups = []Group{{Window: "5m"}, {Window: "15m"}, {Window: "other"}}
	days := [3]map[string]bool{{}, {}, {}}
	if len(archive.Frames) == 0 {
		r.Status = "NO_RECORDED_ORDERBOOK_ARCHIVE"
	}
	for _, p := range positions {
		if (p.Status != "CLOSED" && p.Status != "SETTLED") || p.ExitTime == nil || p.EntryTime.IsZero() || p.Market == "" || !finite(p.RealizedPnL) {
			continue
		}
		idx := 2
		switch windowOf(p.Market) {
		case "5m":
			idx = 0
		case "15m":
			idx = 1
		}
		g := &r.Groups[idx]
		g.RecordedClosedTrades++
		r.AllRecordedClosed++
		audit := AuditEntry(p, archive.Frames, cfg)
		if audit.Status != "OBSERVED_MODEL_ONLY" {
			g.MissingOrUnverifiedEntries++
			r.MissingOrUnverified++
			continue
		}
		g.VerifiedEntrySnapshots++
		if audit.CostOnlyStopBreach {
			g.CostOnlyStopRisk++
		}
		legacy := ReplayExit(p, archive.Frames, cfg, OriginalPolicy)
		alt := ReplayExit(p, archive.Frames, cfg, AlternativePolicy)
		if legacy.ModeledPnLUSD == nil || alt.ModeledPnLUSD == nil {
			r.MissingOrUnverified++
			continue
		}
		g.PairedModeledExits++
		r.MatchedPairs++
		g.RecordedPnLOnPairsUSD += p.RealizedPnL
		g.LegacyModeledPnLUSD += *legacy.ModeledPnLUSD
		g.AlternativeModeledPnLUSD += *alt.ModeledPnLUSD
		days[idx][p.EntryTime.UTC().Format("2006-01-02")] = true
	}
	adequate := true
	for i := range r.Groups {
		g := &r.Groups[i]
		g.IndependentUTCDays = len(days[i])
		g.RecordedPnLOnPairsUSD = cleanMoney(g.RecordedPnLOnPairsUSD)
		g.LegacyModeledPnLUSD = cleanMoney(g.LegacyModeledPnLUSD)
		g.AlternativeModeledPnLUSD = cleanMoney(g.AlternativeModeledPnLUSD)
		if i < 2 && (g.PairedModeledExits < 30 || g.IndependentUTCDays < 10) {
			adequate = false
		}
	}
	if r.ArchiveFrames > 0 && adequate {
		r.Status = "EXPLORATORY_REPLAY_SAMPLE_AVAILABLE_NOT_APPROVED"
	}
	return r
}

// A timestamp from a saved archive is not a current market price. This helper
// deliberately has no network, order signing, or portfolio mutation methods.
