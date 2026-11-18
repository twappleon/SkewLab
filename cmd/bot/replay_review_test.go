package main

import (
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/replay"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReplayReviewPublicAggregateAndFailClosed(t *testing.T) {
	t.Setenv("DASHBOARD_PASSWORD", "secret")
	a := NewApp(t.TempDir(), false, .02, .2, time.Second)
	w := httptest.NewRecorder()
	a.handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/replay-review", nil))
	if w.Code != 200 {
		t.Fatalf("GET %d", w.Code)
	}
	var body replay.Review
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.TradingAuthorized || body.ModelVersion != "1.6.0" || body.ArchiveFrames != 0 {
		t.Fatalf("unexpected report %+v %v", body, err)
	}
	for _, field := range []string{"condition_id", "token_id", "market_slug", "file_path"} {
		if strings.Contains(w.Body.String(), field) {
			t.Fatalf("private field %q", field)
		}
	}
	w = httptest.NewRecorder()
	a.handler().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/replay-review", nil))
	if w.Code != 405 {
		t.Fatalf("POST %d", w.Code)
	}
	w = httptest.NewRecorder()
	a.handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/export.csv", nil))
	if w.Code != 401 {
		t.Fatalf("raw CSV exposed: %d", w.Code)
	}
}
