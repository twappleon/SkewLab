package polymarket

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSettlementGammaStrictResolution(t *testing.T) {
	body := `{"slug":"btc-updown-5m-1800000000","conditionId":"0x123","closed":true,"outcomes":"[\"Down\",\"Up\"]","outcomePrices":"[\"0\",\"1\"]"}`
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/markets/slug/") {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, body)
	}))
	defer s.Close()
	c := New()
	c.Gamma = s.URL
	got, err := c.Settlement(context.Background(), "btc-updown-5m-1800000000", "0x123")
	if err != nil || got.Winner != "UP" {
		t.Fatalf("settlement %+v %v", got, err)
	}
	_, err = c.Settlement(context.Background(), "btc-updown-5m-1800000000", "0xbad")
	if err == nil {
		t.Fatal("accepted condition mismatch")
	}
	body = strings.Replace(body, `"closed":true`, `"closed":false`, 1)
	if _, err = c.Settlement(context.Background(), "btc-updown-5m-1800000000", "0x123"); err == nil {
		t.Fatal("closed=false should block")
	}
	body = strings.Replace(body, `"closed":false`, `"closed":true`, 1)
	body = strings.Replace(body, `[\"0\",\"1\"]`, `[\"0.5\",\"0.5\"]`, 1)
	if _, err = c.Settlement(context.Background(), "btc-updown-5m-1800000000", "0x123"); err == nil {
		t.Fatal("ambiguous prices must block")
	}
}
