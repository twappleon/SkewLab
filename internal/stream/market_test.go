package stream

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestOrderbookSnapshotChangesStalenessAndSorting(t *testing.T) {
	f := NewFeed()
	f.SetAssets([]string{"up", "up", "down"})
	f.mu.Lock()
	f.status.Connected = true
	f.mu.Unlock()
	now := time.Now().UTC()
	ts := fmt.Sprint(now.UnixMilli())
	frame := fmt.Sprintf(`{"event_type":"book","asset_id":"up","timestamp":"%s","tick_size":"0.01","bids":[{"price":"0.42","size":"10"},{"price":"0.40","size":"100"}],"asks":[{"price":"0.49","size":"20"},{"price":"0.47","size":"40"}]}`, ts)
	if e := f.Apply([]byte(frame), now); e != nil {
		t.Fatal(e)
	}
	q, okay := f.Snapshot("up", now)
	if !okay || q.Bid != .42 || q.Ask != .47 || q.TickSize != .01 || q.AskDepth != 40 {
		t.Fatalf("bad quote %+v %t", q, okay)
	}
	change := fmt.Sprintf(`{"event_type":"price_change","timestamp":"%s","price_changes":[{"asset_id":"up","price":"0.42","size":"0","side":"BUY"},{"asset_id":"up","price":"0.48","size":"13","side":"SELL"}]}`, ts)
	if e := f.Apply([]byte(change), now); e != nil {
		t.Fatal(e)
	}
	q, okay = f.Snapshot("up", now)
	if !okay || q.Bid != .4 || q.Ask != .47 {
		t.Fatalf("orderbook delta bad %+v", q)
	}
	if _, ok := f.Snapshot("up", now.Add(13*time.Second)); ok {
		t.Fatal("stale book accepted")
	}
	if _, ok := f.Snapshot("not-subscribed", now); ok {
		t.Fatal("unauthorized asset returned")
	}
	older := fmt.Sprintf(`{"event_type":"book","asset_id":"up","timestamp":"%d","bids":[{"price":"0.88","size":"1"}],"asks":[{"price":"0.90","size":"1"}]}`, now.Add(-time.Second).UnixMilli())
	if e := f.Apply([]byte(older), now); e != nil {
		t.Fatal(e)
	}
	q, _ = f.Snapshot("up", now)
	if q.Bid == .88 {
		t.Fatal("old book overwrote fresh book")
	}
	f.SetAssets([]string{"changed"})
	if _, ok := f.Snapshot("up", now); ok {
		t.Fatal("market rotation retained old book")
	}
}
func TestCanonicalEnvelopeAndLiveTradeChannel(t *testing.T) {
	f := NewFeed()
	f.SetAssets([]string{"abc"})
	f.mu.Lock()
	f.status.Connected = true
	f.mu.Unlock()
	now := time.Now().UTC()
	obj := map[string]any{"topic": "market", "type": "book", "payload": map[string]any{"assetId": "abc", "timestamp": fmt.Sprint(now.UnixMilli()), "tickSize": "0.01", "bids": []any{map[string]string{"price": "0.40", "size": "8"}}, "asks": []any{map[string]string{"price": "0.44", "size": "7"}}}}
	b, _ := json.Marshal(obj)
	if err := f.Apply(b, now); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.Snapshot("abc", now); !ok {
		t.Fatal("canonical snapshot not applied")
	}
	trade := fmt.Sprintf(`{"event_type":"last_trade_price","asset_id":"abc","timestamp":"%d","price":"0.40","size":"4","side":"SELL"}`, now.UnixMilli())
	if err := f.Apply([]byte(trade), now); err != nil {
		t.Fatal(err)
	}
	select {
	case v := <-f.Trades():
		if v.Token != "abc" || v.Size != 4 {
			t.Fatalf("wrong trade %+v", v)
		}
	default:
		t.Fatal("no trade event")
	}
	malformed := strings.Replace(trade, fmt.Sprint(now.UnixMilli()), fmt.Sprint(now.Add(-time.Minute).UnixMilli()), 1)
	if e := f.Apply([]byte(malformed), now); e == nil {
		t.Fatal("stale trade accepted")
	}
}

func TestCanonicalPriceChangeUsesCamelCaseFields(t *testing.T) {
	f := NewFeed()
	f.SetAssets([]string{"X"})
	f.mu.Lock()
	f.status.Connected = true
	f.mu.Unlock()
	now := time.Now().UTC()
	at := fmt.Sprint(now.UnixMilli())
	book := fmt.Sprintf(`{"topic":"market","type":"book","payload":{"assetId":"X","timestamp":"%s","tickSize":"0.01","bids":[{"price":"0.30","size":"5"}],"asks":[{"price":"0.35","size":"7"}]}}`, at)
	if e := f.Apply([]byte(book), now); e != nil {
		t.Fatal(e)
	}
	change := fmt.Sprintf(`{"topic":"market","type":"price_change","payload":{"timestamp":"%s","priceChanges":[{"assetId":"X","price":"0.31","size":"11","side":"BUY"}]}}`, at)
	if e := f.Apply([]byte(change), now); e != nil {
		t.Fatal(e)
	}
	q, ok := f.Snapshot("X", now)
	if !ok || q.Bid != .31 || q.BidDepth != 11 {
		t.Fatalf("camelCase delta ignored: %+v %v", q, ok)
	}
}
