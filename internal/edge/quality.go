package edge

import (
	"fmt"
	"time"
)

// Quality describes the *observed* research archive, not the universe of
// markets that could have been observed. It never authorizes trading.
// Missing candidates cannot be inferred from this archive alone.
type Quality struct {
	Status             string         `json:"status"`
	Total              int            `json:"total"`
	Resolved           int            `json:"resolved"`
	Pending            int            `json:"pending"`
	Overdue            int            `json:"overdue"`
	Invalid            int            `json:"invalid"`
	CapturedLast14Days int            `json:"captured_last_14_days"`
	ResolvedLast14Days int            `json:"resolved_last_14_days"`
	LastCaptureAt      *time.Time     `json:"last_capture_at,omitempty"`
	LastResolutionAt   *time.Time     `json:"last_resolution_at,omitempty"`
	LargestCohort      int            `json:"largest_resolved_cohort"`
	RequiredPerCohort  int            `json:"required_per_cohort"`
	Groups             []QualityGroup `json:"groups"`
	ArchiveError       bool           `json:"archive_error"`
	TradingAuthorized  bool           `json:"trading_authorized"`
	Interpretation     string         `json:"interpretation"`
}

type QualityGroup struct {
	Window             string `json:"window"`
	Captured           int    `json:"captured"`
	Resolved           int    `json:"resolved"`
	Pending            int    `json:"pending"`
	Overdue            int    `json:"overdue"`
	CapturedLast14Days int    `json:"captured_last_14_days"`
}

// AssessQuality flags missing/late labels and stale collection WITHOUT
// creating synthetic outcomes or changing records. Overdue means >30min past
// the market end, not a proven provider outage. Each market is sampled once.
func (r *Recorder) AssessQuality(now time.Time) Quality {
	q := Quality{
		RequiredPerCohort: MinEvidence,
		TradingAuthorized: false,
		Groups:            []QualityGroup{{Window: "5m"}, {Window: "15m"}, {Window: "other"}},
		Interpretation:    "Gamma labels are not independent Chainlink verification. Capture counts cover selected signals only; they cannot measure overall market-feed uptime. No strategy is authorized to trade.",
	}
	if r == nil {
		q.Status = "NO_RESEARCH_RECORDER"
		return q
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.err != nil {
		q.Status = "RESEARCH_ARCHIVE_ERROR"
		q.ArchiveError = true
		return q
	}

	cohorts := make(map[string]int)
	cutoff := now.AddDate(0, 0, -14)
	for _, c := range r.records {
		q.Total++
		group := &q.Groups[2]
		switch c.WindowSeconds {
		case 300:
			group = &q.Groups[0]
		case 900:
			group = &q.Groups[1]
		}
		group.Captured++
		if !c.CapturedAt.After(now) && !c.CapturedAt.Before(cutoff) {
			q.CapturedLast14Days++
			group.CapturedLast14Days++
		}
		if !c.CapturedAt.IsZero() && !c.CapturedAt.After(now) && (q.LastCaptureAt == nil || c.CapturedAt.After(*q.LastCaptureAt)) {
			t := c.CapturedAt
			q.LastCaptureAt = &t
		}
		// Bad chronology and invalid labels are not considered evidence.
		invalid := c.CapturedAt.IsZero() || c.CapturedAt.After(now.Add(3*time.Second)) || c.MarketEnd.Before(c.CapturedAt) || (c.Winner != "" && c.Winner != "UP" && c.Winner != "DOWN") || (c.Winner != "" && (c.ResolvedAt.Before(c.MarketEnd) || c.ResolvedAt.After(now.Add(3*time.Second)))) || (c.Winner == "" && !c.ResolvedAt.IsZero())
		if invalid {
			q.Invalid++
			continue
		}
		if c.Winner == "" {
			q.Pending++
			group.Pending++
			if now.After(c.MarketEnd.Add(30 * time.Minute)) {
				q.Overdue++
				group.Overdue++
			}
			continue
		}
		q.Resolved++
		group.Resolved++
		if !c.ResolvedAt.Before(cutoff) {
			q.ResolvedLast14Days++
		}
		if q.LastResolutionAt == nil || c.ResolvedAt.After(*q.LastResolutionAt) {
			t := c.ResolvedAt
			q.LastResolutionAt = &t
		}
		cohort := fmt.Sprintf("%d/%s/%d", c.WindowSeconds, c.Side, priceBucket(c.Ask))
		cohorts[cohort]++
		if cohorts[cohort] > q.LargestCohort {
			q.LargestCohort = cohorts[cohort]
		}
	}
	// Do not represent a single stale signal as definitive exchange downtime.
	switch {
	case q.Invalid > 0:
		q.Status = "INVALID_RESEARCH_RECORDS"
	case q.Total == 0:
		q.Status = "NO_CANDIDATES_YET"
	case q.Overdue > 0:
		q.Status = "SETTLEMENT_BACKLOG_REVIEW"
	case q.LastCaptureAt != nil && now.Sub(*q.LastCaptureAt) > 24*time.Hour:
		q.Status = "NO_RECENT_QUALIFYING_SIGNALS"
	case q.LargestCohort < MinEvidence:
		q.Status = "COLLECTING_INSUFFICIENT_EVIDENCE"
	default:
		q.Status = "RESEARCH_REVIEW_ONLY"
	}
	return q
}
