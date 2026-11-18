package maker

import (
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/stream"
	"math"
	"sync"
	"time"
)

// QueueStudy is a deliberately one-sided hypothetical passive BUY audit.
// A public trade cannot prove that an unsubmitted order would be filled.
// No quote is sent to Polymarket, no inventory is created and PnL stays zero.
type QueueStudy struct {
	Market                 string    `json:"market"`
	Token                  string    `json:"token"`
	Price                  float64   `json:"bid_price"`
	QueueAheadShares       float64   `json:"initial_visible_queue_ahead_shares"`
	RemainingAhead         float64   `json:"remaining_queue_ahead_shares"`
	HypotheticalSizeShares float64   `json:"hypothetical_order_shares"`
	ObservedTradeShares    float64   `json:"observed_same_price_trade_shares"`
	Candidate              bool      `json:"could_have_filled_if_order_was_posted"`
	Since                  time.Time `json:"since"`
	Status                 string    `json:"status"`
}
type QueueMonitor struct {
	mu             sync.RWMutex
	studies        map[string]QueueStudy
	seen           map[string]time.Time
	candidateCount uint64
}

func NewQueueMonitor() *QueueMonitor {
	return &QueueMonitor{studies: map[string]QueueStudy{}, seen: map[string]time.Time{}}
}
func (q *QueueMonitor) Update(m polymarket.Market, prop Proposal, now time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if prop.Status != "INDICATIVE_ONLY_NO_FILLS" || !m.Active || now.Add(70*time.Second).After(m.End) {
		return
	}
	token, book := m.UpToken, m.Up
	if prop.Side == "DOWN" {
		token, book = m.DownToken, m.Down
	}
	if token == "" || book.Bid <= 0 || book.BidDepth <= 0 || prop.SizeUSD <= 0 {
		return
	}
	prev, exists := q.studies[token]
	if exists && prev.Market == m.Slug && prev.Price == book.Bid {
		return
	}
	shares := prop.SizeUSD / book.Bid
	if !math.IsInf(shares, 0) && shares > 0 && shares < 1000 {
		q.studies[token] = QueueStudy{Market: m.Slug, Token: token, Price: book.Bid, QueueAheadShares: book.BidDepth, RemainingAhead: book.BidDepth, HypotheticalSizeShares: shares, Since: now, Status: "QUEUE_WATCH_ONLY_NO_ORDER"}
	}
	if len(q.studies) > 16 {
		for key, s := range q.studies {
			if s.Since.Before(now.Add(-20 * time.Minute)) {
				delete(q.studies, key)
			}
		}
	}
}

// Observe counts only future public trades at the exact passive bid price.
// Cancellations, price changes, and unrelated trades NEVER reduce queue ahead.
func (q *QueueMonitor) Observe(t stream.Trade) (QueueStudy, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	s, ok := q.studies[t.Token]
	if !ok || !t.Time.After(s.Since) || t.Size <= 0 || t.Price != s.Price || t.Side != "SELL" {
		return QueueStudy{}, false
	}
	// Last-trade stream has no unique trade ID; suppress obvious same-frame
	// duplicates to avoid inflating inferred consumption on reconnect.
	if q.seen[t.Token] == t.Time {
		return QueueStudy{}, false
	}
	q.seen[t.Token] = t.Time
	s.ObservedTradeShares += t.Size
	s.RemainingAhead = math.Max(0, s.RemainingAhead-t.Size)
	if s.ObservedTradeShares >= s.QueueAheadShares+s.HypotheticalSizeShares && !s.Candidate {
		s.Candidate = true
		s.Status = "HYPOTHETICAL_QUEUE_CLEARED_NOT_ACTUAL_FILL"
		q.candidateCount++
	}
	q.studies[t.Token] = s
	return s, true
}
func (q *QueueMonitor) Snapshot() []QueueStudy {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]QueueStudy, 0, len(q.studies))
	for _, s := range q.studies {
		out = append(out, s)
	}
	return out
}
func (q *QueueMonitor) Candidates() uint64 {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.candidateCount
}
