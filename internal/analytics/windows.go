package analytics

import (
	"math"
	"strings"
	"time"

	"github.com/twappleon/SkewLab/internal/paper"
)

// ExitMix counts only completed paper trades and published settlements.
// It does not imply that a hypothetical stop order was executable.
type ExitMix struct {
	TakeProfit       int `json:"take_profit"`
	StopLoss         int `json:"stop_loss"`
	BeforeSettlement int `json:"before_settlement"`
	Settled          int `json:"settled"`
	Other            int `json:"other"`
}

type MarketWindow struct {
	Market      string  `json:"market"`
	Period      Period  `json:"period"`
	Exits       ExitMix `json:"exits"`
	SampleLevel string  `json:"sample_level"`
}

// CompareMarkets groups completed paper trades by 5m, 15m or unknown slugs.
// The last 14 Phnom Penh calendar days include today. Future timestamps,
// unresolved inventory and invalid P&L cannot enter the comparison.
// Only aggregates leave this package, never market IDs or trade IDs.
func CompareMarkets(l *paper.Ledger, now time.Time) []MarketWindow {
	windows := []MarketWindow{
		{Market: "5m", Period: Period{Days: 14}},
		{Market: "15m", Period: Period{Days: 14}},
		{Market: "other", Period: Period{Days: 14}},
	}
	if l == nil {
		for i := range windows {
			windows[i].SampleLevel = "insufficient"
		}
		return windows
	}
	localNow := now.In(phnomPenh)
	today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, phnomPenh)
	start := today.AddDate(0, 0, -13)
	for _, pos := range l.Positions {
		if (pos.Status != "CLOSED" && pos.Status != "SETTLED") || pos.ExitTime == nil ||
			math.IsNaN(pos.RealizedPnL) || math.IsInf(pos.RealizedPnL, 0) {
			continue
		}
		closed := pos.ExitTime.In(phnomPenh)
		if closed.Before(start) || closed.After(localNow) {
			continue
		}
		index := 2
		if strings.HasPrefix(pos.Market, "btc-updown-5m-") {
			index = 0
		}
		if strings.HasPrefix(pos.Market, "btc-updown-15m-") {
			index = 1
		}
		w := &windows[index]
		addMarketPeriod(&w.Period, pos.RealizedPnL)
		switch {
		case pos.Status == "SETTLED":
			w.Exits.Settled++
		case pos.Note == "net take profit":
			w.Exits.TakeProfit++
		case pos.Note == "net stop loss":
			w.Exits.StopLoss++
		case pos.Note == "exit before settlement":
			w.Exits.BeforeSettlement++
		default:
			w.Exits.Other++
		}
	}
	for i := range windows {
		w := &windows[i]
		finish(&w.Period)
		switch {
		case w.Period.Closed < 30:
			w.SampleLevel = "insufficient"
		case w.Period.Closed < 100:
			w.SampleLevel = "preliminary"
		default:
			w.SampleLevel = "research_sample"
		}
	}
	return windows
}

// Independent accounting avoids relying on helper names used by older releases.
func addMarketPeriod(p *Period, pnl float64) {
	p.Closed++
	p.NetPnLUSD += pnl
	if pnl > 0 {
		p.Wins++
		p.GrossProfitUSD += pnl
	}
	if pnl < 0 {
		p.GrossLossUSD -= pnl
	}
}
