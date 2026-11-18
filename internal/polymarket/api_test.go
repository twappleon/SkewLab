package polymarket

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMarketSlug(t *testing.T) {
	tm := time.Unix(1800000123, 0)
	if got := MarketSlug(tm, 300); got != "btc-updown-5m-1800000000" {
		t.Fatal(got)
	}
}
func TestParsedBookUsesBestPriceNotFirstElement(t *testing.T) {
	q, e := parseBook(book{Bids: []level{{"0.32", "10"}, {"0.45", "50"}}, Asks: []level{{"0.55", "23"}, {"0.51", "46"}}})
	if e != nil || q.Bid != 0.45 || q.Ask != 0.51 || q.AskDepth != 46 {
		t.Fatalf("bad book %+v %v", q, e)
	}
}
func TestDiscoverOutcomeTokenMapping(t *testing.T) {
	now := time.Unix(1800000100, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/events" {
			w.Write([]byte(`[{"markets":[{"active":true,"acceptingOrders":true,"closed":false,"conditionId":"0xabc","outcomes":"[\"Down\",\"Up\"]","clobTokenIds":"[\"down-id\",\"up-id\"]"}]}]`))
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()
	c := New()
	c.Gamma = server.URL
	m, err := c.Discover(context.Background(), now, 300)
	if err != nil || m.UpToken != "up-id" || m.DownToken != "down-id" {
		t.Fatalf("incorrect mapping %+v %v", m, err)
	}
}

func TestBookRejectsStaleSnapshot(t *testing.T) {
	timestamp := time.Now().Add(-time.Minute).UnixMilli()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"timestamp":"%d","bids":[{"price":"0.49","size":"50"}],"asks":[{"price":"0.51","size":"50"}]}`, timestamp)
	}))
	defer srv.Close()
	c := New()
	c.CLOB = srv.URL
	if _, err := c.Book(context.Background(), "abc"); err == nil {
		t.Fatal("stale snapshot accepted")
	}
	timestamp = time.Now().UnixMilli()
	if q, err := c.Book(context.Background(), "abc"); err != nil || q.Ask != .51 {
		t.Fatalf("fresh book rejected %+v %v", q, err)
	}
}
