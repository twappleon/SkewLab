# QuantSkew V2 — paper-only preview

This branch adds an opt-in cost-aware gate to the existing historical-observation backtester. It does **not** enable live orders, change the running Railway service, authorize a wallet, prove returns, or automatically send Telegram notifications.

## Fixed risk profile
- Starting paper capital: **100 USDC-equivalent**
- Maximum order and simultaneous exposure: **2 USDC**
- Daily worst-case loss reservation: **2 USDC**, so one open binary position may consume the entire daily budget; losing trades continue to count for that UTC day.
- Existing net take-profit 1%, stop 5%; the 2% daily stop is a *new-entry reservation gate*, not guaranteed exit protection.
- Existing BTC 30-second momentum and 60-second semivariance still apply; venue switches, stale ticks and gaps fail closed.

## Replay
```sh
go test ./internal/quantv2 ./internal/paper ./internal/strategy
go test -race ./...
go vet ./...
go run ./cmd/backtest --strategy quantv2 --file ./data/observations.jsonl
```

QuantSkew V2 samples only recorded observations, checks fresh CLOB bid/ask and depth, estimates both legs' taker fees/slippage, then uses the existing paper ledger. Missing/invalid quotes do not produce fills. A replay is **not** a real-fill validation. The baseline backtest remains available as `--strategy baseline`.

## Acceptance / next gates
- Real recorded replay dataset plus independent walk-forward split not yet supplied or evaluated here.
- Production Dashboard keeps its existing ledger untouched. The private /admin page now shows a read-only V2 candidate count, backed by the authenticated GET /api/quantv2 endpoint. Optional Telegram candidate notices use the existing authorized bot configuration only when QUANTV2_TELEGRAM_ALERTS=1 (disabled by default); each unique market/side is deduplicated by the existing delivery journal. They are labeled READ-ONLY and are not paper fills or performance claims. These changes are not deployed to the running service.
- Must compare baseline/V2 on identical timestamped observations, check true official settlement and maker queue before any deployment approval.
