package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
)

func TestEdgeResearchIsIndependentOfPauseAndPrivate(t *testing.T) {
	t.Setenv("PAPER_ENTRIES_ENABLED", "")
	t.Setenv("DASHBOARD_PASSWORD", "private-test-password")
	a := NewApp(t.TempDir(), false, .02, .2, time.Second)
	now := time.Now().UTC()
	m := polymarket.Market{Slug: "btc-updown-5m-100", ConditionID: "private-condition", Seconds: 300, End: now.Add(4 * time.Minute), Active: true, Up: polymarket.Quote{Ask: .4, Bid: .39, BidDepth: 100, AskDepth: 100, UpdatedAt: now}}
	if yes, err := a.edge.Capture(m, strategy.Result{Direction: "UP", Ready: true}, now); err != nil || !yes {
		t.Fatalf("research candidate not captured: %v %v", yes, err)
	}
	if paperEntriesEnabled() {
		t.Fatal("trading pause unexpectedly lifted")
	}
	srv := httptest.NewServer(a.handler())
	defer srv.Close()
	pub, err := http.Get(srv.URL + "/api/performance")
	if err != nil {
		t.Fatal(err)
	}
	publicJSON, err := io.ReadAll(pub.Body)
	pub.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if pub.StatusCode != 200 || strings.Contains(string(publicJSON), "private-condition") || strings.Contains(string(publicJSON), "btc-updown-5m-100") {
		t.Fatalf("public research leaked market identity: %d", pub.StatusCode)
	}
	var parsed struct {
		Edge struct {
			Total   int `json:"total"`
			Pending int `json:"pending"`
		} `json:"edge_research"`
	}
	if err = json.Unmarshal(publicJSON, &parsed); err != nil || parsed.Edge.Total != 1 || parsed.Edge.Pending != 1 {
		t.Fatalf("bad public aggregate %+v %v", parsed, err)
	}
	private, err := http.Get(srv.URL + "/api/edge-research")
	if err != nil {
		t.Fatal(err)
	}
	private.Body.Close()
	if private.StatusCode != 401 {
		t.Fatalf("unprotected research audit: %d", private.StatusCode)
	}
}
