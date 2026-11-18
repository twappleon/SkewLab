package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/twappleon/SkewLab/internal/wirews"
)

const PolyBoltURL = "wss://ws-live-v2.polymarket.com/ws"

type Credentials struct{ APIKey, Secret, Passphrase, ChainlinkKey string }

func CredentialsFromEnv() Credentials {
	return Credentials{os.Getenv("POLYMARKET_API_KEY"), os.Getenv("POLYMARKET_API_SECRET"), os.Getenv("POLYMARKET_API_PASSPHRASE"), os.Getenv("CHAINLINK_API_KEY")}
}
func (c Credentials) Ready() bool {
	return c.APIKey != "" && c.Secret != "" && c.Passphrase != "" && c.ChainlinkKey != ""
}

type PolyStatus struct {
	Configured    bool      `json:"configured"`
	Connected     bool      `json:"connected"`
	Subscribed    bool      `json:"subscribed"`
	LastTWAP      float64   `json:"last_twap"`
	LastTimestamp time.Time `json:"last_timestamp"`
	LastError     string    `json:"last_error"`
	Samples       int       `json:"samples"`
	Reconnects    uint64    `json:"reconnects"`
	Dropped       uint64    `json:"dropped"`
	Source        string    `json:"source"`
}
type PolyBolt struct {
	mu      sync.RWMutex
	URL     string
	creds   Credentials
	status  PolyStatus
	ticks   []Tick
	lastSeq uint64
	onTick  func(Tick)
}

func NewPolyBolt(c Credentials, save func(Tick)) *PolyBolt {
	return &PolyBolt{URL: PolyBoltURL, creds: c, onTick: save, status: PolyStatus{Configured: c.Ready(), Source: "CHAINLINK TWAP 60s (not connected)"}}
}
func (p *PolyBolt) Status() PolyStatus { p.mu.RLock(); defer p.mu.RUnlock(); return p.status }
func (p *PolyBolt) Ticks() []Tick {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]Tick(nil), p.ticks...)
}

// Seed restores previously captured TWAP ticks without claiming the live
// provider is connected. Historical prices are never used as fresh prices.
func (p *PolyBolt) Seed(records []Tick) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, t := range records {
		if t.Source != "chainlink-twap-60s" || t.Price <= 0 || t.Time.IsZero() {
			continue
		}
		if len(p.ticks) > 0 && !t.Time.After(p.ticks[len(p.ticks)-1].Time) {
			continue
		}
		p.ticks = append(p.ticks, t)
	}
	if len(p.ticks) > 7200 {
		p.ticks = p.ticks[len(p.ticks)-7200:]
	}
	p.status.Samples = len(p.ticks)
	// An archive alone is not a live valid oracle subscription.
}

func (p *PolyBolt) Audit(start time.Time, duration time.Duration) (Audit, error) {
	return Compare(p.Ticks(), start, duration)
}
func (p *PolyBolt) reset(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status.Connected = false
	p.status.Subscribed = false
	p.status.LastTWAP = 0
	p.status.LastTimestamp = time.Time{}
	p.status.Source = "CHAINLINK TWAP 60s (not connected)"
	if err != nil {
		p.status.LastError = err.Error()
	}
	p.lastSeq = 0
	p.status.Reconnects++
}
func (p *PolyBolt) add(t Tick) {
	p.mu.Lock()
	if len(p.ticks) > 0 && t.Time.Before(p.ticks[len(p.ticks)-1].Time) {
		p.mu.Unlock()
		return
	}
	if len(p.ticks) > 0 && t.Time.Equal(p.ticks[len(p.ticks)-1].Time) {
		if t.Price == p.ticks[len(p.ticks)-1].Price {
			p.mu.Unlock()
			return
		}
		p.status.LastError = "conflicting chainlink TWAP timestamp"
		p.mu.Unlock()
		return
	}
	p.ticks = append(p.ticks, t)
	if len(p.ticks) > 7200 {
		p.ticks = p.ticks[len(p.ticks)-7200:]
	}
	p.status.Samples = len(p.ticks)
	p.status.LastTWAP = t.Price
	p.status.LastTimestamp = t.Time
	p.mu.Unlock()
	if p.onTick != nil {
		p.onTick(t)
	}
}

// Apply validates PolyBolt's authenticated channel envelope, provider,
// sequence/drop counters and EXACT 60-second TWAP source. It fails closed.
func (p *PolyBolt) Apply(raw []byte) error {
	var m struct {
		Op       string `json:"op"`
		Code     string `json:"code"`
		Channel  string `json:"channel"`
		Provider string `json:"provider"`
		Seq      uint64 `json:"seq"`
		Dropped  uint64 `json:"dropped"`
		Snapshot bool   `json:"snapshot"`
		Payload  struct {
			Symbol            string      `json:"symbol"`
			WindowSeconds     int         `json:"window_seconds"`
			Timestamp         int64       `json:"timestamp"`
			Value             json.Number `json:"value"`
			FullAccuracyValue string      `json:"full_accuracy_value"`
			Data              []struct {
				Timestamp         int64       `json:"timestamp"`
				Value             json.Number `json:"value"`
				FullAccuracyValue string      `json:"full_accuracy_value"`
			} `json:"data"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	if m.Op == "error" {
		return fmt.Errorf("PolyBolt error: %s", m.Code)
	}
	if m.Op == "subscribed" && m.Channel == "price.crypto.twap" {
		if m.Provider != "" && m.Provider != "chainlink" {
			return errors.New("TWAP provider is not Chainlink")
		}
		p.mu.Lock()
		p.status.Subscribed = true
		p.status.Source = "CHAINLINK TWAP 60s (auth & subscription confirmed)"
		p.mu.Unlock()
		return nil
	}
	if m.Channel != "price.crypto.twap" {
		return nil
	}
	if m.Payload.Symbol != "btcusd" || m.Payload.WindowSeconds != 60 {
		return errors.New("unrecognized TWAP instrument/window")
	}
	p.mu.Lock()
	if !p.status.Subscribed {
		p.mu.Unlock()
		return errors.New("TWAP update before subscribed ack")
	}
	if m.Dropped > 0 {
		p.status.Dropped += m.Dropped
		p.status.LastError = "TWAP sequence gap; reconnect required"
		p.mu.Unlock()
		return errors.New("TWAP dropped frames")
	}
	if m.Seq > 0 && p.lastSeq > 0 && m.Seq != p.lastSeq+1 {
		p.status.LastError = "TWAP sequence gap; reconnect required"
		p.mu.Unlock()
		return errors.New("TWAP sequence gap")
	}
	if m.Seq > 0 {
		p.lastSeq = m.Seq
	}
	p.mu.Unlock()
	parse := func(value json.Number, accurate string) (float64, error) {
		txt := accurate
		if txt == "" {
			txt = string(value)
		}
		v, e := strconv.ParseFloat(txt, 64)
		if e != nil || v <= 0 || math.IsInf(v, 0) || math.IsNaN(v) {
			return 0, errors.New("invalid Chainlink TWAP price")
		}
		return v, nil
	}
	if m.Snapshot {
		for _, point := range m.Payload.Data {
			v, e := parse(point.Value, point.FullAccuracyValue)
			if e != nil {
				return e
			}
			p.add(Tick{Time: time.UnixMilli(point.Timestamp).UTC(), Price: v, Source: "chainlink-twap-60s"})
		}
		return nil
	}
	v, err := parse(m.Payload.Value, m.Payload.FullAccuracyValue)
	if err != nil {
		return err
	}
	at := time.UnixMilli(m.Payload.Timestamp).UTC()
	now := time.Now().UTC()
	if now.Sub(at) > 20*time.Second || at.Sub(now) > 3*time.Second {
		return errors.New("stale Chainlink TWAP update")
	}
	p.add(Tick{Time: at, Price: v, Source: "chainlink-twap-60s"})
	return nil
}
func (p *PolyBolt) Run(ctx context.Context) {
	if !p.creds.Ready() {
		p.mu.Lock()
		p.status.LastError = "NOT CONFIGURED: POLYMARKET_API_KEY/SECRET/PASSPHRASE and CHAINLINK_API_KEY required"
		p.mu.Unlock()
		return
	}
	delay := time.Second
	for ctx.Err() == nil {
		c, err := wirews.Dial(ctx, p.URL)
		if err == nil {
			p.mu.Lock()
			p.status.Connected = true
			p.status.LastError = ""
			p.mu.Unlock()
			auth, _ := json.Marshal(map[string]any{"op": "auth", "rid": "a1", "auth": map[string]string{"apiKey": p.creds.APIKey, "secret": p.creds.Secret, "passphrase": p.creds.Passphrase}, "chainlink_api_key": p.creds.ChainlinkKey})
			err = c.WriteText(auth)
			authed := false
			subSent := false
			for err == nil && ctx.Err() == nil {
				_ = c.SetReadDeadline(time.Now().Add(42 * time.Second))
				msg, e := c.ReadText()
				if e != nil {
					err = e
					break
				}
				var control struct {
					Op   string `json:"op"`
					Code string `json:"code"`
				}
				if json.Unmarshal(msg, &control) != nil {
					continue
				}
				if control.Op == "authed" && !authed {
					authed = true
					sub, _ := json.Marshal(map[string]any{"op": "subscribe", "rid": "s1", "subscriptions": []any{map[string]any{"channel": "price.crypto.twap", "filter": map[string]any{"symbol": "btcusd", "window_seconds": 60, "provider": "chainlink"}}}})
					err = c.WriteText(sub)
					subSent = err == nil
					continue
				}
				if control.Op == "error" {
					err = fmt.Errorf("PolyBolt %s", strings.TrimSpace(control.Code))
					break
				}
				if !subSent {
					continue
				}
				if e := p.Apply(msg); e != nil {
					err = e
					break
				}
			}
			_ = c.Close()
		}
		p.reset(err)
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
