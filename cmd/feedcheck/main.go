// feedcheck is an opt-in, read-only smoke test of live public BTC and
// Polymarket market data. It cannot submit, sign, or simulate orders.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/stream"
)

type MarketCheck struct {
	Seconds    int     `json:"seconds"`
	Slug       string  `json:"slug"`
	Discovered bool    `json:"discovered"`
	REST       bool    `json:"rest_fresh_valid"`
	WS         bool    `json:"ws_fresh_valid"`
	UpBid      float64 `json:"up_bid,omitempty"`
	UpAsk      float64 `json:"up_ask,omitempty"`
	DownBid    float64 `json:"down_bid,omitempty"`
	DownAsk    float64 `json:"down_ask,omitempty"`
	Error      string  `json:"error,omitempty"`
	market     polymarket.Market
}

type Summary struct {
	Version        string        `json:"version"`
	At             time.Time     `json:"at"`
	ReadOnly       bool          `json:"read_only"`
	Successful     bool          `json:"successful"`
	BTCSource      string        `json:"btc_source,omitempty"`
	BTCPrice       float64       `json:"btc_price,omitempty"`
	BTCError       string        `json:"btc_error,omitempty"`
	Markets        []MarketCheck `json:"markets"`
	ConnectedWS    bool          `json:"connected_ws"`
	FreshWSMarket  int           `json:"fresh_ws_market_count"`
	WSReconnects   uint64        `json:"ws_reconnects"`
	WSBookUpdates  uint64        `json:"ws_book_updates"`
	WSError        string        `json:"ws_error,omitempty"`
	ElapsedSeconds float64       `json:"elapsed_seconds"`
	Requirements   string        `json:"requirements"`
}

func validPair(m polymarket.Market, now time.Time) bool {
	if !m.Active || !now.Before(m.End) || m.UpToken == "" || m.DownToken == "" || m.UpToken == m.DownToken {
		return false
	}
	for _, q := range []polymarket.Quote{m.Up, m.Down} {
		if q.Bid <= 0 || q.Ask <= q.Bid || q.Ask >= 1 || q.BidDepth <= 0 || q.AskDepth <= 0 || q.UpdatedAt.IsZero() || now.Sub(q.UpdatedAt) > 12*time.Second || q.UpdatedAt.After(now.Add(3*time.Second)) {
			return false
		}
	}
	return true
}

func verify(ctx context.Context, duration time.Duration) Summary {
	started := time.Now().UTC()
	r := Summary{Version: "0.4.0", At: started, ReadOnly: true, Requirements: "BTC spot + at least one active 5m/15m market with fresh, valid REST and WS Up/Down books"}
	client := polymarket.New()
	spot, err := client.BTC(ctx)
	if err != nil {
		r.BTCError = err.Error()
	} else {
		r.BTCSource = spot.Source
		r.BTCPrice = spot.Price
	}

	assets := []string{}
	for _, secs := range []int{300, 900} {
		m, e := client.Discover(ctx, time.Now().UTC(), secs)
		check := MarketCheck{Seconds: secs, Slug: m.Slug}
		if e != nil {
			check.Error = "Gamma discovery: " + e.Error()
		} else {
			check.Discovered = true
			assets = append(assets, m.UpToken, m.DownToken)
			updated, e2 := client.RefreshQuotes(ctx, m)
			check.market = updated
			if e2 != nil {
				check.Error = "CLOB REST: " + e2.Error()
			} else if !validPair(updated, time.Now().UTC()) {
				check.Error = "CLOB REST: invalid/stale/crossed/empty book"
			} else {
				check.REST = true
				check.UpBid, check.UpAsk = updated.Up.Bid, updated.Up.Ask
				check.DownBid, check.DownAsk = updated.Down.Bid, updated.Down.Ask
			}
		}
		r.Markets = append(r.Markets, check)
	}

	if len(assets) > 0 {
		feed := stream.NewFeed()
		feed.SetAssets(assets)
		wsCtx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()
		go feed.Run(wsCtx)
		ticker := time.NewTicker(350 * time.Millisecond)
		defer ticker.Stop()
		for {
			allReady := true
			for i := range r.Markets {
				m := r.Markets[i].market
				if !r.Markets[i].Discovered {
					continue
				}
				up, okUp := feed.Snapshot(m.UpToken, time.Now().UTC())
				down, okDown := feed.Snapshot(m.DownToken, time.Now().UTC())
				m.Up, m.Down = up, down
				r.Markets[i].WS = okUp && okDown && validPair(m, time.Now().UTC())
				if !r.Markets[i].WS {
					allReady = false
				}
			}
			if allReady {
				break
			}
			select {
			case <-wsCtx.Done():
				goto wsDone
			case <-ticker.C:
			}
		}
	wsDone:
		status := feed.GetStatus()
		r.ConnectedWS = status.Connected
		r.WSReconnects = status.Reconnects
		r.WSBookUpdates = status.BookUpdates
		r.WSError = status.LastError
	}
	for i := range r.Markets {
		if r.Markets[i].REST && r.Markets[i].WS {
			r.FreshWSMarket++
		}
		if r.Markets[i].Discovered && !r.Markets[i].WS {
			if r.Markets[i].Error != "" {
				r.Markets[i].Error += "; "
			}
			r.Markets[i].Error += "WebSocket fresh Up/Down pair not confirmed"
		}
	}
	r.Successful = r.BTCPrice > 0 && r.FreshWSMarket > 0
	r.ElapsedSeconds = time.Since(started).Seconds()
	return r
}

func writeSummary(r Summary, target string) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	if target != "" {
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			return err
		}
		if err := os.WriteFile(target, append(b, '\n'), 0600); err != nil {
			return err
		}
	}
	_, err = fmt.Println(string(b))
	return err
}

func main() {
	timeout := flag.Duration("timeout", 45*time.Second, "maximum duration to wait for live public WS snapshots after REST checks")
	out := flag.String("output", "", "optional local JSON report path (no secrets); printed to stdout either way")
	flag.Parse()
	if *timeout < 2*time.Second || *timeout > 5*time.Minute || strings.TrimSpace(*out) == "-" {
		fmt.Fprintln(os.Stderr, "timeout must be 2s..5m and output must be a filesystem path")
		os.Exit(2)
	}
	// Each provider has its own short HTTP timeout; the context bounds the
	// entire probe, and no connection ever contains wallet credentials.
	ctx, cancel := context.WithTimeout(context.Background(), *timeout+30*time.Second)
	defer cancel()
	r := verify(ctx, *timeout)
	if err := writeSummary(r, *out); err != nil {
		fmt.Fprintln(os.Stderr, "could not write feedcheck report:", err)
		os.Exit(2)
	}
	if !r.Successful {
		os.Exit(1)
	}
}
