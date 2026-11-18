# Polymarket US$100 public scanner — v0.1
Start: `go run ./cmd/polymarket-scan -limit 50`.

This adds independent read-only Gamma market discovery. It DOES NOT modify the existing paused SkewLab strategy, balances, ledger, Railway deployment or Telegram notification service. It contains no trading credential, wallet, private key or order API integration, and never places orders.

Candidates meet only basic market-structure filters, and are **always NO_TRADE**. Gamma top-of-book prices are indicative, not executable CLOB orders. There is no independently calibrated forecast or verified fee schedule/depth/settlement audit. The structural budget parameters are US$100 total, US$5 per position, US$30 total exposure and US$2 daily loss for future paper-only design; they are NOT enforcement of any execution, because this CLI cannot trade.

Run `go test ./internal/polymarketscan ./cmd/polymarket-scan` to check fail-closed behavior.
Further work: CLOB executable depth, cost verification, time-causal out-of-sample calibrated probability forecasts, 30-day prospective paper trading, local legality, and end-user physical-location restrictions. Do not bypass geoblocking with cloud hosting or VPN. Enabling live orders would require a separate explicit request and review.
