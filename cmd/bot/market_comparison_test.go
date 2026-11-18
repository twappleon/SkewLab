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

func TestPublicWindowComparisonAndPrivateLedger(t *testing.T) {
	t.Setenv("DASHBOARD_USER", "admin")
	t.Setenv("DASHBOARD_PASSWORD", "test-secret")
	a := NewApp(t.TempDir(), false, .02, .2, time.Second)
	now := time.Now().UTC().Add(-time.Minute)
	a.ledger.Positions = []paper.Position{
		{ID: "private5", Market: "btc-updown-5m-secret", Token: "secret-token-5", Status: "SETTLED", ExitTime: &now, RealizedPnL: -18},
		{ID: "private15", Market: "btc-updown-15m-secret", Token: "secret-token-15", Status: "CLOSED", ExitTime: &now, RealizedPnL: 2, Note: "net take profit"},
	}
	a.ledger.TotalClosed = 2
	a.ledger.Wins = 1
	a.ledger.RealizedPnL = -16
	server := httptest.NewServer(a.handler())
	defer server.Close()
	resp, err := http.Get(server.URL + "/api/performance")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("public summary %d", resp.StatusCode)
	}
	for _, forbidden := range []string{"private5", "private15", "secret-token", "btc-updown-", "\"positions\":["} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("private identifier exposed: %s", forbidden)
		}
	}
	var v struct {
		MarketWindows []analytics.MarketWindow `json:"market_windows"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatal(err)
	}
	if len(v.MarketWindows) != 3 || v.MarketWindows[0].Period.NetPnLUSD != -18 || v.MarketWindows[1].Period.NetPnLUSD != 2 || v.MarketWindows[2].Period.Closed != 0 {
		t.Fatalf("market windows mismatch: %+v", v.MarketWindows)
	}
	req, err := http.Get(server.URL + "/api/export.csv")
	if err != nil {
		t.Fatal(err)
	}
	req.Body.Close()
	if req.StatusCode != http.StatusUnauthorized {
		t.Fatalf("private ledger exposed %d", req.StatusCode)
	}
}
