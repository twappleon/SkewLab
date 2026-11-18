package paper

import (
	"math"
	"time"
)

// WorstCaseCommittedLossUSD reserves *full stake*, not the optimistic 5% stop.
// Prediction-market contracts can expire worthless. A stop depends on future
// executable bids and liquidity and is NOT a guaranteed loss limit.
// Same-day realized losers consume the daily risk budget; winners do not
// replenish it. Unresolved/open contracts count regardless of entry date.
func (l *Ledger) WorstCaseCommittedLossUSD(now time.Time) float64 {
	if l == nil {
		return math.Inf(1)
	}
	day := now.UTC().Format("2006-01-02")
	total := 0.0
	for _, p := range l.Positions {
		if p.Status == "OPEN" || p.Status == "PENDING_RESOLUTION" {
			if !(p.EntryCost > 0) || math.IsNaN(p.EntryCost) || math.IsInf(p.EntryCost, 0) {
				return math.Inf(1)
			}
			total += p.EntryCost
		} else if p.ExitTime != nil && p.ExitTime.UTC().Format("2006-01-02") == day && p.RealizedPnL < 0 {
			if math.IsNaN(p.RealizedPnL) || math.IsInf(p.RealizedPnL, 0) {
				return math.Inf(1)
			}
			total -= p.RealizedPnL
		}
	}
	return total
}

// CanReserveNewStake checks a conservative pre-entry upper bound. It is
// intentionally fail-closed for invalid values, and does not turn on trading.
func (l *Ledger) CanReserveNewStake(now time.Time) bool {
	if l == nil {
		return false
	}
	stake, cap := l.Config.OrderUSD, l.Config.MaxDailyLossUSD
	if !(stake > 0 && cap > 0) || math.IsNaN(stake) || math.IsInf(stake, 0) || math.IsNaN(cap) || math.IsInf(cap, 0) {
		return false
	}
	reserved := l.WorstCaseCommittedLossUSD(now)
	return !math.IsNaN(reserved) && !math.IsInf(reserved, 0) && reserved+stake <= cap+1e-9
}
