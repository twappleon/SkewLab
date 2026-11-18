package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/twappleon/SkewLab/internal/analytics"
	"github.com/twappleon/SkewLab/internal/paper"
)

func TestDiagnosticsReadOnlyPublicAggregateAndProtectedAudit(t *testing.T) {
	t.Setenv("DASHBOARD_USER", "admin")
	t.Setenv("DASHBOARD_PASSWORD", "test-secret")
	a := NewApp(t.TempDir(), false, .02, .2, time.Second)
	now := time.Now().UTC().Add(-time.Second)
	entry := now.Add(-time.Minute)
	p := paper.Position{ID: "secret-pos-999", Market: "btc-updown-5m-secret-market-id", Token: "secret-token-value", Side: "DOWN", Status: "CLOSED", Note: "net stop loss", EntryTime: entry, ExitTime: &now, Qty: 20, EntryPrice: .5, ExitPrice: .2, EntryCost: 10.25, RealizedPnL: -6.35}
	a.ledger.Positions = []paper.Position{p}
	a.ledger.TotalClosed = 1
	a.ledger.RealizedPnL = -6.35
	srv := httptest.NewServer(a.handler())
	defer srv.Close()
	for _, route := range []string{"/", "/api/performance"} {
		res, e := http.Get(srv.URL + route)
		if e != nil {
			t.Fatal(e)
		}
		body, e := io.ReadAll(res.Body)
		res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}
		if res.StatusCode != 200 {
			t.Fatalf("public %s status %d", route, res.StatusCode)
		}
		for _, secret := range []string{"secret-pos-999", "secret-market-id", "secret-token-value", "\"entry_time\"", "\"side\""} {
			if strings.Contains(string(body), secret) {
				t.Fatalf("public %s leaked %s", route, secret)
			}
		}
		if route == "/api/performance" {
			var got struct {
				Version    string                     `json:"version"`
				RootCauses analytics.RootCauseSummary `json:"root_causes"`
			}
			if err := json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			if got.Version != "1.4.0" || got.RootCauses.CompletedTrades != 1 || got.RootCauses.StopThresholdBreaches != 1 || len(got.RootCauses.ByReasonMarket) != 1 {
				t.Fatalf("public aggregates %+v", got)
			}
		}
	}
	for _, path := range []string{"/api/diagnostics", "/api/state", "/api/export.csv"} {
		res, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != 401 {
			t.Fatalf("anonymous %s returned %d", path, res.StatusCode)
		}
	}
	req, e := http.NewRequest(http.MethodGet, srv.URL+"/api/diagnostics", nil)
	if e != nil {
		t.Fatal(e)
	}
	req.SetBasicAuth("admin", "test-secret")
	res, e := http.DefaultClient.Do(req)
	if e != nil {
		t.Fatal(e)
	}
	body, e := io.ReadAll(res.Body)
	res.Body.Close()
	if e != nil {
		t.Fatal(e)
	}
	if res.StatusCode != 200 || !strings.Contains(string(body), "net_pnl_usd") {
		t.Fatalf("authenticated audit %d: %s", res.StatusCode, body)
	}
	for _, secret := range []string{"secret-pos-999", "secret-market-id", "secret-token-value"} {
		if strings.Contains(string(body), secret) {
			t.Fatalf("audit leaked identifier %s", secret)
		}
	}
}
