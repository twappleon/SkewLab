# SkewLab paper-trading safety pause (2026-11-18)

Observed first completed paper trades: 4, win rate 25%, realized simulated P&L -$38.26. In the loss diagnostics, 3 stop exits exceeded the nominal 5% trigger and $35.96 was beyond those thresholds. These observations prove unsafe risk behavior; the sample is too small to prove BTC momentum always predicts the wrong direction.

## Reproducible no-movement example
At an unchanged 50c bid/ask (zero spread), the simulated entry still pays ask + 0.002 slippage and fees (0.07×p×(1−p) per share plus 25 bps buffer). The immediate simulated sell receives bid − 0.002 and pays sell fees. For a $20 buy this results in roughly 7.97% simulated loss, already worse than the 5% stop trigger. A normal spread increases the loss further.

## Controls
- Default: new live-data paper entries disabled; ONLY explicit `PAPER_ENTRIES_ENABLED=1` can opt in. Keep this unset while uncalibrated.
- Existing positions still use the old exit and settlement logic, preventing orphaned positions.
- No entry if the current same-book round-trip would already trigger the stop or if bid exit depth is missing.
- On Railway the temporary `MIN_MOMENTUM_PCT=999` further blocks entries until review.
- Preserve all historical P&L, without resets or changing past costs.

## Next hypothesis tests
Add real-time entry/exit L1 snapshots, quote-age, liquidity evidence and delay flags. Use 7–14 days of holdout market data to estimate whether an empirically calibrated settlement probability exceeds the executable ask price and true transaction costs. Do not infer future profits from four paper trades.
