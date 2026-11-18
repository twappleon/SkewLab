package oracle

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPolyBoltAuthenticatedSourceAndSequence(t *testing.T) {
	p := NewPolyBolt(Credentials{}, nil)
	if p.Status().Configured {
		t.Fatal("credentials should be required")
	}
	if e := p.Apply([]byte(`{"op":"subscribed","channel":"price.crypto.twap","provider":"binance"}`)); e == nil {
		t.Fatal("must reject nonchainlink twap")
	}
	if e := p.Apply([]byte(`{"op":"subscribed","channel":"price.crypto.twap","provider":"chainlink"}`)); e != nil {
		t.Fatal(e)
	}
	now := time.Now().UTC().Truncate(time.Second)
	snapshot := fmt.Sprintf(`{"channel":"price.crypto.twap","seq":1,"snapshot":true,"payload":{"symbol":"btcusd","window_seconds":60,"data":[{"timestamp":%d,"value":"94200.125"}]}}`, now.Add(-time.Second).UnixMilli())
	if e := p.Apply([]byte(snapshot)); e != nil {
		t.Fatal(e)
	}
	update := fmt.Sprintf(`{"channel":"price.crypto.twap","seq":2,"payload":{"symbol":"btcusd","window_seconds":60,"timestamp":%d,"value":"94203.375"}}`, now.UnixMilli())
	if e := p.Apply([]byte(update)); e != nil {
		t.Fatal(e)
	}
	status := p.Status()
	if status.LastTWAP != 94203.375 || status.Samples != 2 {
		t.Fatalf("bad status %+v", status)
	}
	gap := strings.Replace(update, `"seq":2`, `"seq":5`, 1)
	if e := p.Apply([]byte(gap)); e == nil {
		t.Fatal("undetected drop")
	}
	p.reset(nil)
	if p.Status().LastTWAP != 0 || p.Status().Subscribed {
		t.Fatal("stale oracle price survived reconnect")
	}
}
func TestTWAPNoUnknownSource(t *testing.T) {
	p := NewPolyBolt(Credentials{}, nil)
	_ = p.Apply([]byte(`{"op":"subscribed","channel":"price.crypto.twap","provider":"chainlink"}`))
	msg := fmt.Sprintf(`{"channel":"price.crypto.twap","seq":1,"payload":{"symbol":"ethusd","window_seconds":60,"timestamp":%d,"value":"2"}}`, time.Now().UnixMilli())
	if e := p.Apply([]byte(msg)); e == nil {
		t.Fatal("eth data accepted as BTC")
	}
}
