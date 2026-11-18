package polymarket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const GammaBase = "https://gamma-api.polymarket.com"
const CLOBBase = "https://clob.polymarket.com"

type Quote struct {
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	BidDepth  float64   `json:"bid_depth"`
	AskDepth  float64   `json:"ask_depth"`
	TickSize  float64   `json:"tick_size"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	Source    string    `json:"source,omitempty"`
}
type Market struct {
	Slug        string    `json:"slug"`
	Question    string    `json:"question"`
	ConditionID string    `json:"condition_id"`
	Seconds     int       `json:"seconds"`
	End         time.Time `json:"end"`
	UpToken     string    `json:"up_token"`
	DownToken   string    `json:"down_token"`
	Up          Quote     `json:"up"`
	Down        Quote     `json:"down"`
	Active      bool      `json:"active"`
	Error       string    `json:"error,omitempty"`
}

type Client struct {
	HTTP          *http.Client
	Gamma         string
	CLOB          string
	spotMu        sync.Mutex
	spotPreferred int
}

func New() *Client {
	return &Client{HTTP: &http.Client{Timeout: 6 * time.Second}, Gamma: GammaBase, CLOB: CLOBBase}
}
func (c *Client) get(ctx context.Context, raw string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PolymarketSkewResearch/0.1 (paper-only)")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("http %d from %s", res.StatusCode, req.URL.Host)
	}
	return json.NewDecoder(io.LimitReader(res.Body, 2<<20)).Decode(v)
}
func asStrings(raw json.RawMessage) ([]string, error) {
	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		return values, nil
	}
	var packed string
	if err := json.Unmarshal(raw, &packed); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(packed), &values); err != nil {
		return nil, err
	}
	return values, nil
}

type gm struct {
	Slug            string          `json:"slug"`
	Question        string          `json:"question"`
	ConditionID     string          `json:"conditionId"`
	Outcomes        json.RawMessage `json:"outcomes"`
	Tokens          json.RawMessage `json:"clobTokenIds"`
	EndDate         string          `json:"endDate"`
	Closed          bool            `json:"closed"`
	Active          bool            `json:"active"`
	AcceptingOrders bool            `json:"acceptingOrders"`
}
type event struct {
	Slug    string `json:"slug"`
	Markets []gm   `json:"markets"`
}

func MarketSlug(now time.Time, seconds int) string {
	slot := now.UTC().Unix() / int64(seconds) * int64(seconds)
	return fmt.Sprintf("btc-updown-%dm-%d", seconds/60, slot)
}
func (c *Client) Discover(ctx context.Context, now time.Time, seconds int) (Market, error) {
	if seconds != 300 && seconds != 900 {
		return Market{}, errors.New("only 300s/900s BTC markets supported")
	}
	slug := MarketSlug(now, seconds)
	m := Market{Slug: slug, Seconds: seconds, End: time.Unix(now.Unix()/int64(seconds)*int64(seconds)+int64(seconds), 0).UTC()}
	var evs []event
	if err := c.get(ctx, c.Gamma+"/events?slug="+url.QueryEscape(slug), &evs); err != nil {
		return m, err
	}
	if len(evs) == 0 || len(evs[0].Markets) == 0 {
		return m, fmt.Errorf("market not found: %s", slug)
	}
	g := evs[0].Markets[0]
	m.Question = g.Question
	m.ConditionID = g.ConditionID
	if g.Closed || !g.Active || !g.AcceptingOrders {
		return m, errors.New("market is not accepting orders")
	}
	outcomes, err := asStrings(g.Outcomes)
	if err != nil {
		return m, fmt.Errorf("invalid outcomes: %w", err)
	}
	tokens, err := asStrings(g.Tokens)
	if err != nil {
		return m, fmt.Errorf("invalid token ids: %w", err)
	}
	if len(tokens) != len(outcomes) || len(tokens) != 2 {
		return m, errors.New("unexpected outcome-token mapping")
	}
	for i, name := range outcomes {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "up":
			m.UpToken = tokens[i]
		case "down":
			m.DownToken = tokens[i]
		}
	}
	if m.UpToken == "" || m.DownToken == "" {
		return m, errors.New("market is not a BTC up/down binary market")
	}
	if g.EndDate != "" {
		if t, err := time.Parse(time.RFC3339Nano, g.EndDate); err == nil {
			m.End = t
		}
	}
	if !now.Before(m.End) {
		return m, errors.New("market already expired")
	}
	m.Active = true
	return m, nil
}

type level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}
type book struct {
	Timestamp string  `json:"timestamp"`
	TickSize  string  `json:"tick_size"`
	Bids      []level `json:"bids"`
	Asks      []level `json:"asks"`
}

func parseBook(b book) (Quote, error) {
	q := Quote{}
	if t, err := strconv.ParseFloat(b.TickSize, 64); err == nil && t > 0 && t < 1 {
		q.TickSize = t
	}
	bids := make([]struct{ p, s float64 }, 0, len(b.Bids))
	asks := make([]struct{ p, s float64 }, 0, len(b.Asks))
	for _, l := range b.Bids {
		p, e1 := strconv.ParseFloat(l.Price, 64)
		s, e2 := strconv.ParseFloat(l.Size, 64)
		if e1 == nil && e2 == nil && p > 0 && p < 1 && s > 0 {
			bids = append(bids, struct{ p, s float64 }{p, s})
		}
	}
	for _, l := range b.Asks {
		p, e1 := strconv.ParseFloat(l.Price, 64)
		s, e2 := strconv.ParseFloat(l.Size, 64)
		if e1 == nil && e2 == nil && p > 0 && p < 1 && s > 0 {
			asks = append(asks, struct{ p, s float64 }{p, s})
		}
	}
	sort.Slice(bids, func(i, j int) bool { return bids[i].p > bids[j].p })
	sort.Slice(asks, func(i, j int) bool { return asks[i].p < asks[j].p })
	if len(bids) > 0 {
		q.Bid = bids[0].p
		q.BidDepth = bids[0].s
	}
	if len(asks) > 0 {
		q.Ask = asks[0].p
		q.AskDepth = asks[0].s
	}
	if q.Bid > 0 && q.Ask > 0 && q.Ask < q.Bid {
		return q, errors.New("crossed orderbook")
	}
	return q, nil
}
func (c *Client) Book(ctx context.Context, token string) (Quote, error) {
	if token == "" {
		return Quote{}, errors.New("empty token")
	}
	var b book
	err := c.get(ctx, c.CLOB+"/book?token_id="+url.QueryEscape(token), &b)
	if err != nil {
		return Quote{}, err
	}
	// CLOB snapshots are not guaranteed fresh simply because an HTTP request
	// succeeded; reject stale, future, or missing update timestamps.
	ms, conv := strconv.ParseInt(b.Timestamp, 10, 64)
	if conv != nil || ms <= 0 {
		return Quote{}, errors.New("CLOB snapshot missing/invalid timestamp")
	}
	at := time.UnixMilli(ms)
	now := time.Now().UTC()
	if now.Sub(at) > 12*time.Second || at.Sub(now) > 3*time.Second {
		return Quote{}, fmt.Errorf("stale/future CLOB snapshot timestamp: %s", at.Format(time.RFC3339Nano))
	}
	q, parseErr := parseBook(b)
	q.UpdatedAt = at
	q.Source = "CLOB_REST"
	return q, parseErr
}
func (c *Client) RefreshQuotes(ctx context.Context, m Market) (Market, error) {
	up, e1 := c.Book(ctx, m.UpToken)
	down, e2 := c.Book(ctx, m.DownToken)
	m.Up = up
	m.Down = down
	if e1 != nil {
		return m, fmt.Errorf("up book: %w", e1)
	}
	if e2 != nil {
		return m, fmt.Errorf("down book: %w", e2)
	}
	return m, nil
}
