package validation

import (
	"testing"
	"time"
)

func TestPerformanceReviewHasFiveRoundDiagnosticsAndNeverApprovesEntry(t *testing.T) {
	now := time.Date(2026, 11, 18, 13, 0, 0, 0, time.UTC)
	r := BuildReview(nil, now)
	if r.ModelVersion != "1.5.0" || r.PaperEntriesAuthorized || r.LiveTradingAuthorized || len(r.Friction.Windows) != 2 || len(r.Regimes.Windows) != 2 || len(r.Incremental.Windows) != 2 {
		t.Fatalf("new fields not present: %+v", r)
	}
	seen := map[string]bool{}
	for _, b := range r.Blockers {
		seen[b] = true
	}
	for _, key := range []string{"NO_VALID_FIXED_TIME_MARKET_OBSERVATIONS", "NO_VALIDATED_INCREMENTAL_EDGE_5m", "NO_VALIDATED_INCREMENTAL_EDGE_15m", "INSUFFICIENT_UTC_DAY_DIVERSITY_5m"} {
		if !seen[key] {
			t.Fatalf("missing conservative blocker %s", key)
		}
	}
}
