package main

import (
	"bytes"
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/validation"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicStudyReviewIsAggregatedFailClosed(t *testing.T) {
	t.Setenv("DASHBOARD_PASSWORD", "test-only-secret")
	t.Setenv("PAPER_ENTRIES_ENABLED", "")
	a := NewApp(t.TempDir(), false, .02, .2, time.Second)
	w := httptest.NewRecorder()
	a.handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/study-review", nil))
	if w.Code != 200 {
		t.Fatalf("public review status %d", w.Code)
	}
	var v validation.Review
	if e := json.Unmarshal(w.Body.Bytes(), &v); e != nil || v.PaperEntriesAuthorized || v.LiveTradingAuthorized || v.Status != "RESEARCH_BLOCKED" {
		t.Fatalf("bad review %+v %v", v, e)
	}
	for _, secret := range [][]byte{[]byte("condition_id"), []byte("token_id"), []byte("market_slug")} {
		if bytes.Contains(w.Body.Bytes(), secret) {
			t.Fatalf("raw data leaked %q", secret)
		}
	}
	w = httptest.NewRecorder()
	a.handler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/study-review", nil))
	if w.Code != 405 {
		t.Fatal("POST must be disabled")
	}
	w = httptest.NewRecorder()
	a.handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/strategy-comparison", nil))
	if w.Code != 401 {
		t.Fatal("private archive must remain protected")
	}
}
