// Package fees models the Polymarket DeFi taker fee schedule for BTC
// five- and fifteen-minute markets. No taker or maker rebate is assumed.
package fees

import (
	"math"
	"time"
)

var ChangeUTC = time.Date(2026, 11, 20, 23, 59, 0, 0, time.UTC)

// CryptoCoefficient is 0.07 before the 2026-11-20 change and 0.05 after.
// Applies only to the BTC 5m/15m Up/Down markets on Polymarket DeFi.
func CryptoCoefficient(at time.Time) float64 {
	if !at.UTC().Before(ChangeUTC) {
		return .05
	}
	return .07
}

func TakerFee(shares, price float64, at time.Time) float64 {
	if shares <= 0 || price <= 0 || price >= 1 || math.IsNaN(price) || math.IsInf(price, 0) || math.IsNaN(shares) || math.IsInf(shares, 0) {
		return 0
	}
	return shares * CryptoCoefficient(at) * price * (1 - price)
}

// BufferedCost adds an extra configurable *conservative simulation reserve*;
// it is NOT a claim that the platform charges this reserve.
func BufferedCost(shares, price, extraBps float64, at time.Time) float64 {
	return shares*price + TakerFee(shares, price, at) + shares*price*extraBps/10000
}
func BufferedProceeds(shares, price, extraBps float64, at time.Time) float64 {
	return shares*price - TakerFee(shares, price, at) - shares*price*extraBps/10000
}
