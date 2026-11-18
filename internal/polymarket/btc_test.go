package polymarket

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type fakeHTTP func(*http.Request) (*http.Response, error)

func (f fakeHTTP) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }
func respond(code int, body string) *http.Response {
	return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}
}

func TestSpotFallbackStickyAndFailureReasons(t *testing.T) {
	counts := map[string]int{}
	c := New()
	c.HTTP = &http.Client{Transport: fakeHTTP(func(req *http.Request) (*http.Response, error) {
		counts[req.URL.Host]++
		switch req.URL.Host {
		case "api.binance.com":
			return respond(451, `{"message":"restricted"}`), nil
		case "api.exchange.coinbase.com":
			return respond(503, `{}`), nil
		case "api.kraken.com":
			return respond(200, `{"error":[],"result":{"XXBTZUSD":{"c":["95432.10","0.0001"]}}}`), nil
		default:
			return respond(500, `{}`), nil
		}
	})}
	first, e := c.BTC(context.Background())
	if e != nil || first.Price != 95432.10 || !strings.Contains(first.Source, "Kraken") {
		t.Fatalf("bad Kraken fallback: %+v %v", first, e)
	}
	if counts["api.binance.com"] != 1 || counts["api.exchange.coinbase.com"] != 1 {
		t.Fatalf("missing initial failed venues: %v", counts)
	}
	second, e := c.BTC(context.Background())
	if e != nil || second.Source != first.Source || counts["api.binance.com"] != 1 {
		t.Fatalf("did not stick to Kraken: %+v %v count=%v", second, e, counts)
	}
}

func TestSpotBitstampFallbackAndStaleRefusal(t *testing.T) {
	c := New()
	fresh := fmt.Sprint(time.Now().Unix())
	c.HTTP = &http.Client{Transport: fakeHTTP(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "www.bitstamp.net" {
			return respond(200, fmt.Sprintf(`{"last":"93111.5","timestamp":"%s"}`, fresh)), nil
		}
		return respond(451, `{}`), nil
	})}
	got, e := c.BTC(context.Background())
	if e != nil || got.Source != "Bitstamp BTCUSD spot" {
		t.Fatalf("bitstamp fallback: %+v %v", got, e)
	}
	c = New()
	c.HTTP = &http.Client{Transport: fakeHTTP(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host == "www.bitstamp.net" {
			return respond(200, `{"last":"99999.0","timestamp":"1"}`), nil
		}
		return respond(403, `{}`), nil
	})}
	_, e = c.BTC(context.Background())
	if e == nil || !strings.Contains(e.Error(), "Bitstamp") || !strings.Contains(e.Error(), "Binance") {
		t.Fatalf("should report all unavailable venues: %v", e)
	}
}
func TestSpotMalformedDoesNotCount(t *testing.T) {
	for _, s := range []string{"", "NaN", "Inf", "-1", "0", "bad"} {
		if _, err := validSpot(s); err == nil {
			t.Fatalf("accepted %q", s)
		}
	}
}
