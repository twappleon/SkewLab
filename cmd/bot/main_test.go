package main

import (
	"context"
	"encoding/json"
	"github.com/twappleon/SkewLab/internal/paper"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDemoEndToEndAndExport(t *testing.T) {
	dir := t.TempDir()
	a := NewApp(dir, true, .02, .2, 3*time.Second)
	a.cycle(context.Background())
	s := httptest.NewServer(a.handler())
	defer s.Close()
	resp, err := http.Get(s.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var state struct {
		Mode string `json:"mode"`
		BTC  struct {
			Source string `json:"source"`
		}
		Markets []struct{ Active bool }
		Ledger  struct{ Positions []any }
	}
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		t.Fatal(err)
	}
	if state.Mode != "SYNTHETIC DEMO" || state.BTC.Source != "SYNTHETIC DEMO (not live)" || len(state.Markets) != 2 {
		t.Fatalf("demo broken %+v", state)
	}
	if _, err := os.Stat(filepath.Join(dir, "observations.jsonl")); err != nil {
		t.Fatal(err)
	}
	resp2, err := http.Get(s.URL + "/api/export.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Fatal(resp2.StatusCode)
	}
}
func TestPrivateDashboardAuth(t *testing.T) {
	t.Setenv("DASHBOARD_USER", "operator")
	t.Setenv("DASHBOARD_PASSWORD", "secret")
	a := NewApp(t.TempDir(), true, .02, .2, time.Second)
	server := httptest.NewServer(a.handler())
	defer server.Close()
	resp, err := http.Get(server.URL + "/api/state")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("must reject unauthenticated API, got %d", resp.StatusCode)
	}
	req, _ := http.NewRequest("GET", server.URL+"/api/state", nil)
	req.SetBasicAuth("operator", "secret")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("good login rejected: %d", resp.StatusCode)
	}
	resp, err = http.Get(server.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("health should be public for container checks")
	}
}

func TestReconcileOfficialOutcomeOnceWithoutSpot(t *testing.T) {
	now := time.Now().UTC()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/markets/slug/closed-test" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"slug":"closed-test","conditionId":"0xdef","closed":true,"outcomes":"[\"Up\",\"Down\"]","outcomePrices":"[\"1\",\"0\"]"}`))
	}))
	defer server.Close()
	a := NewApp(t.TempDir(), false, .02, .2, time.Second)
	a.api.Gamma = server.URL
	a.ledger.Cash = 980
	a.ledger.Positions = []paper.Position{{Market: "closed-test", ConditionID: "0xdef", Side: "UP", Status: "PENDING_RESOLUTION", Qty: 40, EntryCost: 20, MarketEnd: now.Add(-time.Minute), MarkPnL: -20}}
	a.reconcile(context.Background(), now)
	if a.ledger.Positions[0].Status != "SETTLED" || a.ledger.Cash != 1020 || a.ledger.RealizedPnL != 20 {
		t.Fatalf("not reconciled %+v", a.ledger)
	}
	a.reconcile(context.Background(), now.Add(time.Minute))
	if a.ledger.Cash != 1020 || a.ledger.TotalClosed != 1 {
		t.Fatal("duplicate settlement")
	}
}
