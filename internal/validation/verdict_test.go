package validation

import (
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/experiment"
	"strings"
	"testing"
	"time"
)

func TestReviewFailClosedPrivacyAndNoInventedSamples(t *testing.T) {
	now := time.Date(2026, 11, 18, 11, 0, 0, 0, time.UTC)
	r := BuildReview(nil, now)
	if r.Status != "RESEARCH_BLOCKED" || len(r.Blockers) == 0 || r.PaperEntriesAuthorized || r.LiveTradingAuthorized || !r.StudyOnly {
		t.Fatal(r)
	}
	s := fixture(now.Add(-time.Hour), 300, "UP")
	s.Slug = "private-market-id"
	s.ConditionID = "private-condition-id"
	r = BuildReview([]experiment.Sample{s}, now)
	b, e := json.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(b), s.Slug) || strings.Contains(string(b), s.ConditionID) {
		t.Fatal("private identifiers leaked")
	}
	if !strings.Contains(string(b), "INSUFFICIENT_FORWARD_MATCHED") || r.PaperEntriesAuthorized {
		t.Fatal(r)
	}
}
