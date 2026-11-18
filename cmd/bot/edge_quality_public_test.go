package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/twappleon/SkewLab/internal/edge"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
)

func TestPublicEvidenceQualityOnlyAggregateAndNoTrading(t *testing.T) {
	t.Setenv("PAPER_ENTRIES_ENABLED", "")
	t.Setenv("DASHBOARD_PASSWORD", "secret-for-tests")
	a := NewApp(t.TempDir(), false, .02, .2, time.Second)
	now := time.Now().UTC()
	m := polymarket.Market{Slug: "btc-updown-5m-private", ConditionID: "PRIVATE-CONDITION", Seconds: 300, End: now.Add(4 * time.Minute), Active: true, Up: polymarket.Quote{Ask: .42, Bid: .41, AskDepth: 100, BidDepth: 100, UpdatedAt: now}}
	if yes, err := a.edge.Capture(m, strategy.Result{Ready: true, Direction: "UP", MomentumPct: .04, Skew: .4}, now); !yes || err != nil {
		t.Fatalf("capture %v %v", yes, err)
	}
	srv := httptest.NewServer(a.handler())
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/api/performance")
	if err != nil {
		t.Fatal(err)
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("aggregate API status %d", resp.StatusCode)
	}
	for _, secret := range []string{"PRIVATE-CONDITION", "btc-updown-5m-private", "\"recent_private_samples\""} {
		if strings.Contains(string(body), secret) {
			t.Fatalf("private content leaked: %s", secret)
		}
	}
	var public struct {
		Quality edge.Quality `json:"edge_quality"`
	}
	if err := json.Unmarshal(body, &public); err != nil {
		t.Fatal(err)
	}
	if public.Quality.Total != 1 || public.Quality.Groups[0].Captured != 1 || public.Quality.TradingAuthorized || public.Quality.Status != "COLLECTING_INSUFFICIENT_EVIDENCE" {
		t.Fatalf("bad quality: %+v", public.Quality)
	}
	r, err := http.Get(srv.URL + "/api/edge-research")
	if err != nil {
		t.Fatal(err)
	}
	r.Body.Close()
	if r.StatusCode != 401 {
		t.Fatalf("private research exposed %d", r.StatusCode)
	}
}
