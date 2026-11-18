package validation

import (
	"github.com/twappleon/SkewLab/internal/experiment"
	"time"
)

type Review struct {
	AsOf                   time.Time        `json:"as_of"`
	Status                 string           `json:"status"`
	StudyOnly              bool             `json:"study_only"`
	PaperEntriesAuthorized bool             `json:"paper_entries_authorized"`
	LiveTradingAuthorized  bool             `json:"live_trading_authorized"`
	Blockers               []string         `json:"blockers"`
	Integrity              Integrity        `json:"integrity"`
	Settlement             SettlementHealth `json:"settlement"`
	Paired                 PairedReport     `json:"paired"`
	Stress                 StressReport     `json:"stress"`
	Disclosure             string           `json:"disclosure"`
}

// BuildReview never returns a trading authorization even when all exploratory
// checks pass. Separate early-exit execution and user approval are required.
func BuildReview(samples []experiment.Sample, now time.Time) Review {
	r := Review{AsOf: now, StudyOnly: true, Status: "RESEARCH_BLOCKED", PaperEntriesAuthorized: false, LiveTradingAuthorized: false,
		Disclosure: "Read-only, prospective, modeled hold-to-settlement research. Gamma outcomes are not independent Chainlink evidence. No executable order, early stop or real-money conclusion is implied."}
	r.Integrity = AuditIntegrity(samples, now)
	r.Settlement = SettlementAudit(samples, now)
	r.Paired = PairedCheck(samples, now, .002, 25)
	r.Stress = StressCompare(samples, now)
	if r.Integrity.Valid == 0 {
		r.Blockers = append(r.Blockers, "NO_VALID_FIXED_TIME_MARKET_OBSERVATIONS")
	}
	if r.Integrity.Invalid > 0 {
		r.Blockers = append(r.Blockers, "INVALID_OR_DUPLICATE_OBSERVATIONS")
	}
	if r.Settlement.Overdue > 0 {
		r.Blockers = append(r.Blockers, "UNRESOLVED_MARKETS_OVERDUE")
	}
	for _, w := range []string{"5m", "15m"} {
		for _, x := range r.Paired.ByWindow[w] {
			if x.MatchedSettled < 30 || x.IndependentUTCDays < 10 {
				r.Blockers = append(r.Blockers, "INSUFFICIENT_FORWARD_MATCHED_"+w+"_"+x.Rule)
			}
			if x.ConservativeDayBlockLowerUSD == nil || *x.ConservativeDayBlockLowerUSD <= 0 {
				r.Blockers = append(r.Blockers, "NO_POSITIVE_DAY_BLOCK_LOWER_BOUND_"+w+"_"+x.Rule)
			}
		}
	}
	if !r.Stress.AllScenariosPositive {
		r.Blockers = append(r.Blockers, "COST_STRESS_NOT_PROVEN")
	}
	if len(r.Blockers) == 0 {
		r.Status = "RESEARCH_REVIEW_REQUIRED_NO_TRADING_APPROVAL"
	}
	return r
}
