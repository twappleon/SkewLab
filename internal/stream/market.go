// Package stream maintains a public read-only Polymarket CLOB market feed.
// It rejects stale messages and never sends trading/order instructions.
package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/wirews"
)

const PublicMarketWS = "wss://ws-subscriptions-clob.polymarket.com/ws/market"

type Level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}
type PriceChange struct {
	AssetID string `json:"asset_id"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Side    string `json:"side"`
	BestBid string `json:"best_bid"`
	BestAsk string `json:"best_ask"`
}
type rawEvent struct {
	EventType    string          `json:"event_type"`
	Type         string          `json:"type"`
	AssetID      string          `json:"asset_id"`
	Timestamp    string          `json:"timestamp"`
	Hash         string          `json:"hash"`
	TickSize     string          `json:"tick_size"`
	Bids         []Level         `json:"bids"`
	Asks         []Level         `json:"asks"`
	PriceChanges []PriceChange   `json:"price_changes"`
	Price        string          `json:"price"`
	Size         string          `json:"size"`
	Side         string          `json:"side"`
	Payload      json.RawMessage `json:"payload"`
}
type Trade struct {
	Token string    `json:"token"`
	Price float64   `json:"price"`
	Size  float64   `json:"size"`
	Side  string    `json:"side"`
	Time  time.Time `json:"time"`
}
type BookState struct {
	Bids         map[float64]float64
	Asks         map[float64]float64
	Updated      time.Time
	TickSize     float64
	Hash         string
	HaveSnapshot bool
}
type Status struct {
	Connected          bool      `json:"connected"`
	LastMessage        time.Time `json:"last_message"`
	LastError          string    `json:"last_error"`
	Reconnects         uint64    `json:"reconnects"`
	BookUpdates        uint64    `json:"book_updates"`
	TradeEvents        uint64    `json:"trade_events"`
	DroppedTradeEvents uint64    `json:"dropped_trade_events"`
	Subscribed         int       `json:"subscribed_tokens"`
	Source             string    `json:"source"`
}

// Feed is safe for concurrent reads and one Run / one Apply writer.
type Feed struct {
	mu     sync.RWMutex
	wanted []string
	books  map[string]*BookState
	status Status
	active *wirews.Conn
	trades chan Trade
	URL    string
}

func NewFeed() *Feed {
	return &Feed{books: make(map[string]*BookState), trades: make(chan Trade, 512), URL: PublicMarketWS, status: Status{Source: "PUBLIC CLOB WS (unverified until connected)"}}
}
func (f *Feed) Trades() <-chan Trade { return f.trades }
func unique(tokens []string) []string {
	m := map[string]bool{}
	for _, v := range tokens {
		if v != "" {
			m[v] = true
		}
	}
	out := make([]string, 0, len(m))
	for t := range m {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
func (f *Feed) SetAssets(tokens []string) {
	tokens = unique(tokens)
	f.mu.Lock()
	defer f.mu.Unlock()
	if strings.Join(tokens, "|") == strings.Join(f.wanted, "|") {
		return
	}
	f.wanted = tokens
	f.books = make(map[string]*BookState)
	f.status.Subscribed = len(tokens)
	if f.active != nil {
		_ = f.active.Close()
	}
}
func (f *Feed) Snapshot(token string, now time.Time) (polymarket.Quote, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	b := f.books[token]
	if !f.status.Connected || b == nil || !b.HaveSnapshot || now.Sub(b.Updated) > 8*time.Second || b.Updated.After(now.Add(3*time.Second)) {
		return polymarket.Quote{}, false
	}
	p, s := best(b.Bids, true)
	a, z := best(b.Asks, false)
	if p <= 0 || a <= p || s <= 0 || z <= 0 {
		return polymarket.Quote{}, false
	}
	return polymarket.Quote{Bid: p, Ask: a, BidDepth: s, AskDepth: z, TickSize: b.TickSize, Source: "CLOB_WEBSOCKET", UpdatedAt: b.Updated}, true
}
func best(book map[float64]float64, bid bool) (float64, float64) {
	price := 0.0
	sz := 0.0
	for p, s := range book {
		if s <= 0 {
			continue
		}
		if price == 0 || (bid && p > price) || (!bid && p < price) {
			price, sz = p, s
		}
	}
	return price, sz
}
func (f *Feed) GetStatus() Status { f.mu.RLock(); defer f.mu.RUnlock(); return f.status }
func parseNum(v string) float64 {
	n, e := strconv.ParseFloat(v, 64)
	if e != nil || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0
	}
	return n
}
func applyLevels(ls []Level) map[float64]float64 {
	m := map[float64]float64{}
	for _, l := range ls {
		p, s := parseNum(l.Price), parseNum(l.Size)
		if p > 0 && p < 1 && s > 0 && len(m) < 10000 {
			m[p] = s
		}
	}
	return m
}
func timeFromMillis(ms string) (time.Time, error) {
	v, e := strconv.ParseInt(ms, 10, 64)
	if e != nil || v <= 0 {
		return time.Time{}, errors.New("missing/invalid feed timestamp")
	}
	return time.UnixMilli(v).UTC(), nil
}

// Apply accepts both legacy CLOB envelopes and SDK canonical payload events.
// No orderbook updates or trade events are accepted for unrequested tokens.
func (f *Feed) Apply(raw []byte, received time.Time) error {
	if len(raw) == 0 || string(raw) == "PING" || string(raw) == "PONG" {
		return nil
	}
	if raw[0] == '[' {
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			return err
		}
		for _, it := range items {
			if err := f.Apply(it, received); err != nil {
				return err
			}
		}
		return nil
	}
	var ev rawEvent
	if err := json.Unmarshal(raw, &ev); err != nil {
		return err
	}
	if ev.Type == "" {
		ev.Type = ev.EventType
	}
	if ev.Type == "" {
		return nil
	}
	if len(ev.Payload) > 0 {
		var p struct {
			AssetID      string                                                          `json:"assetId"`
			Timestamp    string                                                          `json:"timestamp"`
			Hash         string                                                          `json:"hash"`
			TickSize     string                                                          `json:"tickSize"`
			Bids         []Level                                                         `json:"bids"`
			Asks         []Level                                                         `json:"asks"`
			PriceChanges []struct{ AssetID, Price, Size, Side, BestBid, BestAsk string } `json:"priceChanges"`
			Price        string                                                          `json:"price"`
			Size         string                                                          `json:"size"`
			Side         string                                                          `json:"side"`
		}
		if e := json.Unmarshal(ev.Payload, &p); e != nil {
			return e
		}
		ev.AssetID = p.AssetID
		ev.Bids = p.Bids
		ev.Asks = p.Asks
		ev.Hash = p.Hash
		ev.TickSize = p.TickSize
		ev.Price = p.Price
		ev.Size = p.Size
		ev.Side = p.Side
		ev.Timestamp = p.Timestamp
		for _, change := range p.PriceChanges {
			ev.PriceChanges = append(ev.PriceChanges, PriceChange{AssetID: change.AssetID, Price: change.Price, Size: change.Size, Side: change.Side, BestBid: change.BestBid, BestAsk: change.BestAsk})
		}
	}
	at, err := timeFromMillis(ev.Timestamp)
	if err != nil {
		return err
	}
	if received.Sub(at) > 12*time.Second || at.Sub(received) > 3*time.Second {
		return errors.New("market websocket message stale or future")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	allowed := func(t string) bool {
		for _, a := range f.wanted {
			if a == t {
				return true
			}
		}
		return false
	}
	f.status.LastMessage = received
	if ev.Type == "book" {
		if !allowed(ev.AssetID) {
			return nil
		}
		prev := f.books[ev.AssetID]
		if prev != nil && at.Before(prev.Updated) {
			return nil
		}
		b := &BookState{Bids: applyLevels(ev.Bids), Asks: applyLevels(ev.Asks), Updated: at, TickSize: parseNum(ev.TickSize), Hash: ev.Hash, HaveSnapshot: true}
		if b.TickSize <= 0 && prev != nil {
			b.TickSize = prev.TickSize
		}
		f.books[ev.AssetID] = b
		f.status.BookUpdates++
		return nil
	}
	if ev.Type == "price_change" {
		for _, change := range ev.PriceChanges {
			if !allowed(change.AssetID) {
				continue
			}
			b := f.books[change.AssetID]
			if b == nil || !b.HaveSnapshot || at.Before(b.Updated) {
				continue
			}
			price, size := parseNum(change.Price), parseNum(change.Size)
			if price <= 0 || price >= 1 {
				continue
			}
			var lv map[float64]float64
			switch strings.ToUpper(change.Side) {
			case "BUY":
				lv = b.Bids
			case "SELL":
				lv = b.Asks
			default:
				continue
			}
			if size <= 0 {
				delete(lv, price)
			} else {
				lv[price] = size
			}
			b.Updated = at
			f.status.BookUpdates++
		}
		return nil
	}
	if ev.Type == "tick_size_change" {
		if !allowed(ev.AssetID) {
			return nil
		}
		b := f.books[ev.AssetID]
		if b != nil && at.After(b.Updated) {
			b.TickSize = parseNum(ev.TickSize)
		}
		return nil
	}
	if ev.Type == "last_trade_price" && allowed(ev.AssetID) {
		p, s := parseNum(ev.Price), parseNum(ev.Size)
		if p <= 0 || p >= 1 || s <= 0 {
			return nil
		}
		f.status.TradeEvents++
		select {
		case f.trades <- Trade{Token: ev.AssetID, Price: p, Size: s, Side: strings.ToUpper(ev.Side), Time: at}:
		default:
			f.status.DroppedTradeEvents++
		}
	}
	return nil
}

// Run reconnects on failed reads, stale streams, and rolling market subscriptions.
func (f *Feed) Run(ctx context.Context) {
	delay := time.Second
	for ctx.Err() == nil {
		f.mu.RLock()
		assets := append([]string(nil), f.wanted...)
		target := f.URL
		f.mu.RUnlock()
		if len(assets) == 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				continue
			}
		}
		c, err := wirews.Dial(ctx, target)
		if err == nil {
			f.mu.RLock()
			same := strings.Join(assets, "|") == strings.Join(f.wanted, "|")
			f.mu.RUnlock()
			if !same {
				_ = c.Close()
				continue
			}
			f.mu.Lock()
			f.active = c
			f.status.Connected = true
			f.status.LastError = ""
			f.mu.Unlock()
			subscribe, _ := json.Marshal(struct {
				Assets []string `json:"assets_ids"`
				Type   string   `json:"type"`
				Custom bool     `json:"custom_feature_enabled"`
			}{assets, "market", true})
			err = c.WriteText(subscribe)
			if err == nil {
				// Public CLOB market stream requires a client PING every 10s.
				// A separate writer avoids waiting for incoming messages; all
				// writes are masked and serialized by wirews.Conn.
				heartbeatDone := make(chan struct{})
				go func() {
					tick := time.NewTicker(10 * time.Second)
					defer tick.Stop()
					for {
						select {
						case <-ctx.Done():
							_ = c.Close()
							return
						case <-heartbeatDone:
							return
						case <-tick.C:
							if e := c.WriteText([]byte("PING")); e != nil {
								_ = c.Close()
								return
							}
						}
					}
				}()
				delay = time.Second
				for ctx.Err() == nil {
					_ = c.SetReadDeadline(time.Now().Add(35 * time.Second))
					var msg []byte
					msg, err = c.ReadText()
					if err != nil {
						break
					}
					if string(msg) == "PING" {
						err = c.WriteText([]byte("PONG"))
						if err != nil {
							break
						}
						continue
					}
					if e := f.Apply(msg, time.Now().UTC()); e != nil && os.Getenv("SKEWLAB_VERBOSE_STREAM") == "1" {
						fmt.Printf("WS invalid event: %s\n", e)
					}
				}
				close(heartbeatDone)
			}
			_ = c.Close()
		}
		f.mu.Lock()
		if f.active == c {
			f.active = nil
		}
		f.status.Connected = false
		f.books = make(map[string]*BookState)
		f.status.Reconnects++
		if err != nil {
			f.status.LastError = err.Error()
		}
		f.mu.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		if delay < 16*time.Second {
			delay *= 2
		}
	}
}
