package replay

import (
	"github.com/twappleon/SkewLab/internal/fees"
	"github.com/twappleon/SkewLab/internal/paper"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"math"
	"sort"
	"time"
)

// EntryAudit uses a timestamped snapshot as a *model consistency check*, not
// proof that a real order or even a hypothetical fill occurred.
type EntryAudit struct {
	Status                    string     `json:"status"`
	ObservedAt                *time.Time `json:"observed_at,omitempty"`
	RoundTripLossPct          *float64   `json:"same_quote_round_trip_loss_pct,omitempty"`
	CostOnlyStopBreach        bool       `json:"cost_only_stop_breach"`
	QuoteMatchesRecordedEntry bool       `json:"quote_matches_recorded_entry"`
}

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func quoteFor(m polymarket.Market, side string) polymarket.Quote {
	if side == "UP" {
		return m.Up
	}
	if side == "DOWN" {
		return m.Down
	}
	return polymarket.Quote{}
}
func marketFor(frame Frame, pos paper.Position) (polymarket.Market, bool) {
	for _, m := range frame.Markets {
		if m.Slug == pos.Market && (pos.ConditionID == "" || m.ConditionID == pos.ConditionID) {
			return m, true
		}
	}
	return polymarket.Market{}, false
}

// validQuote rejects bad prices and depths before any replay inference.
// A quote timestamp AFTER the observation time would leak future information.
func validQuote(q polymarket.Quote, at time.Time) bool {
	if !finite(q.Bid) || !finite(q.Ask) || !finite(q.BidDepth) || !finite(q.AskDepth) {
		return false
	}
	if q.Bid <= 0 || q.Bid >= 1 || q.Ask <= 0 || q.Ask >= 1 || q.Bid > q.Ask || q.BidDepth <= 0 || q.AskDepth <= 0 {
		return false
	}
	if at.IsZero() || q.UpdatedAt.IsZero() || q.UpdatedAt.After(at) || at.Sub(q.UpdatedAt) > 12*time.Second {
		return false
	}
	return true
}

// AuditEntry rejects post-entry observations and never invents a missing book.
func AuditEntry(pos paper.Position, frames []Frame, cfg paper.Config) EntryAudit {
	out := EntryAudit{Status: "NO_PRE_ENTRY_SNAPSHOT"}
	if pos.EntryTime.IsZero() || !(pos.Qty > 0) || !(pos.EntryCost > 0) || !finite(pos.EntryCost) || !finite(pos.Qty) || pos.Market == "" || !(cfg.StopLossPct > 0) || cfg.SlippagePrice < 0 {
		return EntryAudit{Status: "INVALID_POSITION_OR_CONFIG"}
	}
	i := sort.Search(len(frames), func(i int) bool { return frames[i].Time.After(pos.EntryTime) }) - 1
	for ; i >= 0 && pos.EntryTime.Sub(frames[i].Time) <= 8*time.Second; i-- {
		f := frames[i]
		if f.Time.After(pos.EntryTime) {
			continue
		}
		m, ok := marketFor(f, pos)
		if !ok {
			continue
		}
		q := quoteFor(m, pos.Side)
		at := pos.EntryTime
		if !validQuote(q, at) || !validQuote(q, f.Time) {
			return EntryAudit{Status: "STALE_OR_INVALID_ENTRY_BOOK"}
		}
		if q.AskDepth+1e-6 < pos.Qty || q.BidDepth+1e-6 < pos.Qty {
			return EntryAudit{Status: "INSUFFICIENT_VISIBLE_DEPTH"}
		}
		out.Status = "OBSERVED_MODEL_ONLY"
		t := f.Time
		out.ObservedAt = &t
		out.QuoteMatchesRecordedEntry = math.Abs(pos.EntryPrice-(q.Ask+cfg.SlippagePrice)) <= 0.003
		bid := q.Bid - cfg.SlippagePrice
		if !(bid > 0 && bid < 1) {
			out.Status = "INVALID_HYPOTHETICAL_EXIT"
			return out
		}
		proceeds := fees.BufferedProceeds(pos.Qty, bid, cfg.FeeBps, pos.EntryTime)
		if !finite(proceeds) {
			out.Status = "INVALID_HYPOTHETICAL_EXIT"
			return out
		}
		loss := 100 * (pos.EntryCost - proceeds) / pos.EntryCost
		out.RoundTripLossPct = &loss
		out.CostOnlyStopBreach = loss >= cfg.StopLossPct
		if !out.QuoteMatchesRecordedEntry {
			out.Status = "QUOTE_MISMATCH_NO_ATTRIBUTION"
		}
		return out
	}
	return out
}
