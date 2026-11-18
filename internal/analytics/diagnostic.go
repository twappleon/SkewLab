package analytics

import (
	"math"
	"sort"
	"strings"
	"time"

	"github.com/twappleon/SkewLab/internal/paper"
)

// TradeAttribution separates the recorded change in contract price from the
// costs embedded in the paper ledger. It is NOT an estimate of real fills,
// bid/ask spread or actual on-chain fees. Entry/exit prices already include
// the assumed price slippage and cannot reveal the independent effect of
// actual spread or latency. No source tick reconstruction is attempted.
type TradeAttribution struct {
	MarketGroup           string    `json:"market_group"`
	ExitReason            string    `json:"exit_reason"`
	EntryTime             time.Time `json:"entry_time"`
	ExitTime              time.Time `json:"exit_time"`
	EntryCostUSD          float64   `json:"entry_cost_usd"`
	NetPnLUSD             float64   `json:"net_pnl_usd"`
	NetPnLPct             float64   `json:"net_pnl_pct"`
	ContractPricePnLUSD   *float64  `json:"contract_price_pnl_usd"`
	ModeledCostReserveUSD *float64  `json:"modeled_cost_reserve_usd"`
	StopLossThresholdPct  float64   `json:"stop_loss_threshold_pct"`
	ExceededStopThreshold bool      `json:"exceeded_stop_threshold"`
	LossBeyondStopUSD     float64   `json:"loss_beyond_stop_usd"`
	HoldingSeconds        *float64  `json:"holding_seconds"`
	AttributionComplete   bool      `json:"attribution_complete"`
}

type LossBucket struct {
	Reason                string  `json:"reason"`
	Market                string  `json:"market"`
	Trades                int     `json:"trades"`
	NetPnLUSD             float64 `json:"net_pnl_usd"`
	GrossLossUSD          float64 `json:"gross_loss_usd"`
	GrossProfitUSD        float64 `json:"gross_profit_usd"`
	StopThresholdBreaches int     `json:"stop_threshold_breaches"`
	ExcessStopLossUSD     float64 `json:"excess_stop_loss_usd"`
}

type RootCauseSummary struct {
	WindowDays            int          `json:"window_days"`
	CompletedTrades       int          `json:"completed_trades"`
	AnalyzedTrades        int          `json:"analyzed_trades"`
	MissingAttribution    int          `json:"missing_attribution"`
	CostReserveUSD        float64      `json:"modeled_cost_reserve_usd"`
	PriceMovementPnLUSD   float64      `json:"contract_price_pnl_usd"`
	NetPnLUSD             float64      `json:"net_pnl_usd"`
	StopThresholdBreaches int          `json:"stop_threshold_breaches"`
	ExcessStopLossUSD     float64      `json:"excess_stop_loss_usd"`
	ByReasonMarket        []LossBucket `json:"by_reason_market"`
	AttributionCaveat     string       `json:"attribution_caveat"`
}

func diagnosisMarket(s string) string {
	if strings.HasPrefix(s, "btc-updown-5m-") {
		return "5m"
	}
	if strings.HasPrefix(s, "btc-updown-15m-") {
		return "15m"
	}
	return "other"
}
func diagnosisReason(pos paper.Position) string {
	if pos.Status == "SETTLED" {
		return "official_settlement"
	}
	switch pos.Note {
	case "net stop loss":
		return "stop_loss"
	case "net take profit":
		return "take_profit"
	case "exit before settlement":
		return "before_settlement"
	default:
		return "other"
	}
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func maxZero(v float64) float64 {
	if v > 0 {
		return v
	}
	return 0
}

// ExplainCompleted deliberately omits IDs, raw market slugs, token IDs, side
// and source addresses. Per-trade data should be behind the authenticated API;
// public views use only SummarizeCauses aggregated over many trades.
func ExplainCompleted(p paper.Position, stopPct float64) (TradeAttribution, bool) {
	d := TradeAttribution{MarketGroup: diagnosisMarket(p.Market), ExitReason: diagnosisReason(p), EntryTime: p.EntryTime, EntryCostUSD: p.EntryCost, StopLossThresholdPct: stopPct}
	if (p.Status != "CLOSED" && p.Status != "SETTLED") || p.ExitTime == nil || p.EntryCost <= 0 || !finite(p.EntryCost) || !finite(p.RealizedPnL) {
		return d, false
	}
	d.ExitTime = *p.ExitTime
	d.NetPnLUSD = round(p.RealizedPnL)
	d.NetPnLPct = round(p.RealizedPnL / p.EntryCost * 100)
	d.ExceededStopThreshold = p.Note == "net stop loss" && stopPct > 0 && d.NetPnLPct < -stopPct-0.00001
	if d.ExceededStopThreshold {
		d.LossBeyondStopUSD = round(maxZero(-p.RealizedPnL - p.EntryCost*stopPct/100))
	}
	if !p.EntryTime.IsZero() && !p.ExitTime.Before(p.EntryTime) {
		elapsed := round(p.ExitTime.Sub(p.EntryTime).Seconds())
		d.HoldingSeconds = &elapsed
	}
	// A priced, valid complete paper trade admits an identity:
	// realized P&L = qty*(exit-entry) - (entry+exit cost reserves).
	// For Gamma settlement, exit price is 0 or 1 and no exit fee is booked.
	if p.Qty > 0 && finite(p.Qty) && finite(p.EntryPrice) && finite(p.ExitPrice) && p.EntryPrice >= 0 && p.EntryPrice <= 1 && p.ExitPrice >= 0 && p.ExitPrice <= 1 {
		entryReserve := p.EntryCost - p.Qty*p.EntryPrice
		pricePNL := p.Qty * (p.ExitPrice - p.EntryPrice)
		reserve := pricePNL - p.RealizedPnL
		if entryReserve >= -0.0001 && reserve >= -0.0001 && finite(pricePNL) && finite(reserve) {
			reserve = math.Max(reserve, 0)
			a, b := round(pricePNL), round(reserve)
			d.ContractPricePnLUSD = &a
			d.ModeledCostReserveUSD = &b
			d.AttributionComplete = true
		}
	}
	return d, true
}

func SummarizeCauses(l *paper.Ledger, now time.Time) RootCauseSummary {
	out := RootCauseSummary{WindowDays: 14, AttributionCaveat: "僅依模擬帳本的進出價與成本做會計分解；不能證明是行情跳動、買賣價差、滑價或訂單延遲造成。止損是觸發條件，不是保證最大損失。"}
	if l == nil {
		out.ByReasonMarket = []LossBucket{}
		return out
	}
	localNow := now.In(phnomPenh)
	day := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, phnomPenh)
	start := day.AddDate(0, 0, -13)
	buckets := map[string]*LossBucket{}
	for _, p := range l.Positions {
		if p.ExitTime == nil || (p.Status != "CLOSED" && p.Status != "SETTLED") {
			continue
		}
		date := p.ExitTime.In(phnomPenh)
		if date.Before(start) || date.After(localNow) {
			continue
		}
		d, ok := ExplainCompleted(p, l.Config.StopLossPct)
		if !ok {
			out.MissingAttribution++
			continue
		}
		out.CompletedTrades++
		out.NetPnLUSD += p.RealizedPnL
		key := d.MarketGroup + "/" + d.ExitReason
		bucket := buckets[key]
		if bucket == nil {
			bucket = &LossBucket{Market: d.MarketGroup, Reason: d.ExitReason}
			buckets[key] = bucket
		}
		bucket.Trades++
		bucket.NetPnLUSD += p.RealizedPnL
		if p.RealizedPnL < 0 {
			bucket.GrossLossUSD -= p.RealizedPnL
		} else if p.RealizedPnL > 0 {
			bucket.GrossProfitUSD += p.RealizedPnL
		}
		if d.ExceededStopThreshold {
			bucket.StopThresholdBreaches++
			bucket.ExcessStopLossUSD += d.LossBeyondStopUSD
			out.StopThresholdBreaches++
			out.ExcessStopLossUSD += d.LossBeyondStopUSD
		}
		if d.AttributionComplete {
			out.AnalyzedTrades++
			out.PriceMovementPnLUSD += *d.ContractPricePnLUSD
			out.CostReserveUSD += *d.ModeledCostReserveUSD
		} else {
			out.MissingAttribution++
		}
	}
	for _, v := range buckets {
		v.NetPnLUSD = round(v.NetPnLUSD)
		v.GrossLossUSD = round(v.GrossLossUSD)
		v.GrossProfitUSD = round(v.GrossProfitUSD)
		v.ExcessStopLossUSD = round(v.ExcessStopLossUSD)
		out.ByReasonMarket = append(out.ByReasonMarket, *v)
	}
	sort.Slice(out.ByReasonMarket, func(i, j int) bool {
		a, b := out.ByReasonMarket[i], out.ByReasonMarket[j]
		if a.Market != b.Market {
			return a.Market < b.Market
		}
		return a.Reason < b.Reason
	})
	out.NetPnLUSD = round(out.NetPnLUSD)
	out.CostReserveUSD = round(out.CostReserveUSD)
	out.PriceMovementPnLUSD = round(out.PriceMovementPnLUSD)
	out.ExcessStopLossUSD = round(out.ExcessStopLossUSD)
	return out
}

// PrivateDiagnostics returns an audit trail without issuing orders or exposing
// Telegram secrets. All callers must be authenticated by the HTTP handler.
func PrivateDiagnostics(l *paper.Ledger, now time.Time, limit int) []TradeAttribution {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	out := []TradeAttribution{}
	if l == nil {
		return out
	}
	for _, p := range l.Positions {
		d, ok := ExplainCompleted(p, l.Config.StopLossPct)
		if !ok || d.ExitTime.After(now) {
			continue
		}
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ExitTime.After(out[j].ExitTime) })
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
