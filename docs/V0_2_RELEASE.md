# SkewLab v0.2.0 — release notes

## Delivered

- Gamma official closed market settlement reconciliation with matching slug, condition ID, terminal binary prices, and idempotency.
- Failed/unresolved settlement cannot be guessed from Binance, Coinbase, or indicative market bid/ask.
- Expired paper inventory conservatively marked zero until official result is available.
- Date-aware Polymarket BTC 5/15-minute taker fee curve, plus explicit configurable conservative cost buffer, no assumed rebates.
- Actual CLOB tick-size-aware maker research proposals, **no simulated maker fills**.
- Timestamp freshness validation on CLOB HTTP book snapshots.
- Chainlink TWAP **offline tick audit** (requires actual externally captured exact Chainlink boundary ticks; no live feed implementation).
- Dashboard, RWD, unit tests, offline integration tests, CSV and JSONL data paths updated.

## Pass/fail evidence

- `go test ./...`, `go test -race ./...`, `go vet ./...`, `go build` — verified locally.
- Synthetic demo HTTP `/health` and `/api/state` — verified locally.
- Browser desktop and 390px mobile via headless Chromium with a captured synthetic JSON response: no JS errors or horizontal document overflow.
- Live Gamma/CLOB/BTC calls — NOT tested: local DNS blocked.
- Cloud deployment — NOT done; separate Sprite creation blocked by exhausted trial billing.
- Authenticated Chainlink/PolyBolt subscriptions — NOT implemented; credentials not supplied.
- Real-money execution — NOT implemented and disabled by architecture.

This release DOES NOT establish positive expected value or live-market operational reliability.
