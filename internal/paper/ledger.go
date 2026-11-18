package paper

import (
	"github.com/twappleon/SkewLab/internal/fees"
	"github.com/twappleon/SkewLab/internal/polymarket"
	"github.com/twappleon/SkewLab/internal/strategy"
	"math"
	"time"
)

type Config struct {
	StartingCash    float64 `json:"starting_cash"`
	OrderUSD        float64 `json:"order_usd"`
	MaxExposureUSD  float64 `json:"max_exposure_usd"`
	MaxDailyLossUSD float64 `json:"max_daily_loss_usd"`
	MaxDailyEntries int     `json:"max_daily_entries"`
	FeeBps          float64 `json:"assumed_fee_bps_per_fill"` // additional conservative fee reserve beyond official per-share fee
	SlippagePrice   float64 `json:"assumed_slippage_price_per_fill"`
	ProfitTargetPct float64 `json:"net_profit_target_pct"`
	StopLossPct     float64 `json:"net_stop_loss_pct"`
	MaxBuyPrice     float64 `json:"max_buy_price"`
	MinDepthShares  float64 `json:"min_top_of_book_depth_shares"`
	CooldownSeconds int     `json:"cooldown_seconds"`
}

func DefaultConfig() Config {
	return Config{StartingCash: 1000, OrderUSD: 20, MaxExposureUSD: 80, MaxDailyLossUSD: 30, MaxDailyEntries: 35, FeeBps: 25, SlippagePrice: 0.002, ProfitTargetPct: 1, StopLossPct: 5, MaxBuyPrice: 0.75, MinDepthShares: 20, CooldownSeconds: 35}
}

type Position struct {
	ID          string     `json:"id"`
	Market      string     `json:"market"`
	ConditionID string     `json:"condition_id,omitempty"`
	Side        string     `json:"side"`
	Token       string     `json:"token"`
	Qty         float64    `json:"qty"`
	EntryPrice  float64    `json:"entry_price"`
	EntryCost   float64    `json:"entry_cost"`
	EntryTime   time.Time  `json:"entry_time"`
	MarketEnd   time.Time  `json:"market_end"`
	ExitPrice   float64    `json:"exit_price"`
	ExitTime    *time.Time `json:"exit_time,omitempty"`
	Status      string     `json:"status"`
	RealizedPnL float64    `json:"realized_pnl"`
	MarkPnL     float64    `json:"mark_pnl"`
	Note        string     `json:"note"`
}

type EquityPoint struct {
	Time   time.Time `json:"time"`
	Equity float64   `json:"equity"`
}
type Ledger struct {
	Config         Config        `json:"config"`
	Cash           float64       `json:"cash"`
	Positions      []Position    `json:"positions"`
	Equity         []EquityPoint `json:"equity"`
	LastEntry      time.Time     `json:"last_entry"`
	LastEvaluation time.Time     `json:"last_evaluation"`
	TotalClosed    int           `json:"total_closed"`
	Wins           int           `json:"wins"`
	RealizedPnL    float64       `json:"realized_pnl"`
	HighWatermark  float64       `json:"high_watermark"`
	MaxDrawdownPct float64       `json:"max_drawdown_pct"`
}

func New(c Config) *Ledger {
	return &Ledger{Config: c, Cash: c.StartingCash, HighWatermark: c.StartingCash, Equity: []EquityPoint{}}
}
func (l *Ledger) exposure() float64 {
	n := 0.0
	for _, p := range l.Positions {
		if p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION" {
			n += p.EntryCost
		}
	}
	return n
}
func (l *Ledger) todayLoss(now time.Time) float64 {
	val := 0.0
	for _, p := range l.Positions {
		if p.ExitTime != nil && p.ExitTime.UTC().Format("2006-01-02") == now.UTC().Format("2006-01-02") {
			val += p.RealizedPnL
		} else if (p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION") && p.MarkPnL < 0 {
			val += p.MarkPnL
		}
	}
	return val
}
func (l *Ledger) entriesToday(now time.Time) int {
	n := 0
	for _, p := range l.Positions {
		if p.EntryTime.UTC().Format("2006-01-02") == now.UTC().Format("2006-01-02") {
			n++
		}
	}
	return n
}
func (l *Ledger) quote(m polymarket.Market, side string) polymarket.Quote {
	if side == "UP" {
		return m.Up
	}
	return m.Down
}

// ExpireUnresolved removes the provisional value of expired inventory even
// when *all* spot and CLOB data sources are offline. No settlement is guessed.
func (l *Ledger) ExpireUnresolved(now time.Time) int {
	changed := 0
	for i := range l.Positions {
		p := &l.Positions[i]
		if p.Status == "OPEN" && !now.Before(p.MarketEnd) {
			p.Status = "PENDING_RESOLUTION"
			p.MarkPnL = -p.EntryCost
			p.Note = "unresolved: conservative zero-value pending official outcome"
			changed++
		}
	}
	return changed
}

func (l *Ledger) mark(m polymarket.Market, now time.Time) {
	l.ExpireUnresolved(now)
	for i := range l.Positions {
		p := &l.Positions[i]
		if p.Status != "OPEN" {
			continue
		}
		if !now.Before(p.MarketEnd) {
			continue
		}
		if p.Market != m.Slug || !m.Active {
			continue
		}
		q := l.quote(m, p.Side)
		if q.Bid > 0 {
			exPrice := math.Max(0, q.Bid-l.Config.SlippagePrice)
			net := fees.BufferedProceeds(p.Qty, exPrice, l.Config.FeeBps, now)
			p.MarkPnL = net - p.EntryCost
		}
	}
}
func (l *Ledger) Evaluate(m polymarket.Market, s strategy.Result, now time.Time) []string {
	messages := []string{}
	l.mark(m, now)
	if m.Active {
		for i := range l.Positions {
			p := &l.Positions[i]
			if p.Status != "OPEN" || p.Market != m.Slug {
				continue
			}
			q := l.quote(m, p.Side)
			if q.Bid <= 0 || q.BidDepth < p.Qty {
				continue
			} // no invented fills
			pnlPct := 100 * p.MarkPnL / p.EntryCost
			nearEnd := now.Add(32 * time.Second).After(p.MarketEnd)
			if now.Sub(p.EntryTime) < 4*time.Second && !nearEnd {
				continue
			}
			if pnlPct >= l.Config.ProfitTargetPct || pnlPct <= -l.Config.StopLossPct || nearEnd {
				price := math.Max(0.001, q.Bid-l.Config.SlippagePrice)
				proceeds := fees.BufferedProceeds(p.Qty, price, l.Config.FeeBps, now)
				p.ExitPrice = price
				p.ExitTime = &now
				p.Status = "CLOSED"
				p.RealizedPnL = proceeds - p.EntryCost
				p.MarkPnL = 0
				if nearEnd {
					p.Note = "exit before settlement"
				} else if pnlPct >= l.Config.ProfitTargetPct {
					p.Note = "net take profit"
				} else {
					p.Note = "net stop loss"
				}
				l.Cash += proceeds
				l.RealizedPnL += p.RealizedPnL
				l.TotalClosed++
				if p.RealizedPnL > 0 {
					l.Wins++
				}
				messages = append(messages, "PAPER SELL "+p.Side+" "+m.Slug)
			}
		}
	}
	if s.Direction != "UP" && s.Direction != "DOWN" {
		return messages
	}
	if !s.Ready || !m.Active || now.Add(65*time.Second).After(m.End) {
		return messages
	}
	if !l.CanReserveNewStake(now) {
		return messages
	}
	if l.todayLoss(now) <= -l.Config.MaxDailyLossUSD || l.entriesToday(now) >= l.Config.MaxDailyEntries {
		return messages
	}
	if now.Sub(l.LastEntry) < time.Duration(l.Config.CooldownSeconds)*time.Second {
		return messages
	}
	if l.exposure()+l.Config.OrderUSD > l.Config.MaxExposureUSD {
		return messages
	}
	for _, p := range l.Positions {
		if p.Market == m.Slug && (p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION") {
			return messages
		}
	}
	q := l.quote(m, s.Direction)
	px := q.Ask + l.Config.SlippagePrice
	if q.Ask <= 0 || px > l.Config.MaxBuyPrice || px >= 1 || q.Ask < q.Bid {
		return messages
	}
	qty := l.Config.OrderUSD / fees.BufferedCost(1, px, l.Config.FeeBps, now)
	if q.AskDepth < math.Max(qty, l.Config.MinDepthShares) {
		return messages
	}
	// Prevent cost-only stop-loss churn: even an unchanged bid/ask can
	// generate a loss past the stop after round-trip taker fees and slippage.
	// A successful cost gate is NOT evidence of a statistical trading edge.
	if q.Bid <= 0 || q.BidDepth < qty || l.Config.StopLossPct <= 0 {
		return messages
	}
	exitPx := math.Max(0.001, q.Bid-l.Config.SlippagePrice)
	sameBookProceeds := fees.BufferedProceeds(qty, exitPx, l.Config.FeeBps, now)
	frictionPct := 100 * (l.Config.OrderUSD-sameBookProceeds) / l.Config.OrderUSD
	if math.IsNaN(frictionPct) || math.IsInf(frictionPct, 0) || frictionPct >= l.Config.StopLossPct {
		return messages
	}
	if l.Cash < l.Config.OrderUSD {
		return messages
	}
	token := m.UpToken
	if s.Direction == "DOWN" {
		token = m.DownToken
	}
	p := Position{ID: now.UTC().Format("20060102T150405.000000000") + "-" + s.Direction, Market: m.Slug, ConditionID: m.ConditionID, Side: s.Direction, Token: token, Qty: qty, EntryPrice: px, EntryCost: l.Config.OrderUSD, EntryTime: now, MarketEnd: m.End, Status: "OPEN", Note: "paper taker fill at best ask + configured slippage"}
	l.Positions = append(l.Positions, p)
	l.Cash -= p.EntryCost
	l.LastEntry = now
	messages = append(messages, "PAPER BUY "+p.Side+" "+m.Slug)
	return messages
}
func (l *Ledger) RecordEquity(now time.Time) {
	equity := l.Cash
	for _, p := range l.Positions {
		if p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION" {
			equity += p.EntryCost + p.MarkPnL
		}
	}
	if equity > l.HighWatermark {
		l.HighWatermark = equity
	}
	if l.HighWatermark > 0 {
		dd := 100 * (l.HighWatermark - equity) / l.HighWatermark
		if dd > l.MaxDrawdownPct {
			l.MaxDrawdownPct = dd
		}
	}
	l.LastEvaluation = now
	l.Equity = append(l.Equity, EquityPoint{Time: now, Equity: equity})
	if len(l.Equity) > 600 {
		l.Equity = l.Equity[len(l.Equity)-600:]
	}
}
func (l *Ledger) CurrentEquity() float64 {
	if len(l.Equity) > 0 {
		return l.Equity[len(l.Equity)-1].Equity
	}
	return l.Cash
}
func (l *Ledger) OpenPositions() int {
	c := 0
	for _, p := range l.Positions {
		if p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION" {
			c++
		}
	}
	return c
}

// ApplySettlement reconciles expired paper inventory against the officially
// published Polymarket market result, never a Coinbase/Binance spot tick.
// Settlement is idempotent; already-closed positions are not modified.
func (l *Ledger) ApplySettlement(result polymarket.Settlement, now time.Time) int {
	count := 0
	for i := range l.Positions {
		p := &l.Positions[i]
		if p.Market != result.Slug || p.ConditionID == "" || p.ConditionID != result.ConditionID ||
			(p.Status != "OPEN" && p.Status != "PENDING_RESOLUTION") || now.Before(p.MarketEnd) {
			continue
		}
		proceeds := 0.0
		if p.Side == result.Winner {
			proceeds = p.Qty
		}
		p.ExitPrice = 0
		if proceeds > 0 {
			p.ExitPrice = 1
		}
		p.ExitTime = &now
		p.RealizedPnL = proceeds - p.EntryCost
		p.MarkPnL = 0
		p.Status = "SETTLED"
		p.Note = "official Polymarket Gamma reported winner; not independently verified against Chainlink"
		l.Cash += proceeds
		l.RealizedPnL += p.RealizedPnL
		l.TotalClosed++
		if p.RealizedPnL > 0 {
			l.Wins++
		}
		count++
	}
	return count
}
