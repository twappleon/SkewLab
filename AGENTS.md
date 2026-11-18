# Agent instructions — SkewLab Polymarket BTC Quant

## Mission
Deliver a **scientifically testable**, transparent, fail-closed BTC prediction-market research and paper-trading system. Never equate a profitable screenshot, synthetic demo, or CI success with verified profitability.

## Non-negotiable gates

1. No real orders, signer keys, or wallet integrations without explicit user consent, legal/geoblocking checks, tested kill switches, and an extended approved paper result.
2. Never invent BTC/Chainlink, bid/ask, trades, performance, market outcomes, timestamps, or public URLs.
3. The same reference venue must feed a whole momentum and semivariance calculation. Reset after feed switch/disconnect. Reject observations older than 5 seconds or with gaps >8 seconds.
4. Use *executable CLOB ask* for simulated buy and *executable CLOB bid* for simulated sale; account for realistic fees, slippage, depth, and unresolved settlement. Distinguish hypothetical fills from actual fills.
5. Net exit profit target remains **at least +1% after all simulated costs**; loss controls and maximum exposure are mandatory. Profits are never guaranteed.
6. Every feature requires unit tests, replay tests, and, when available, read-only real-market validation. Record versions and precise limitations in README.
7. Never deploy a trading dashboard publicly without explicit access control. No secrets in source or logs.
8. Respect legitimate platform restrictions, local laws, and Polymarket official documentation; no geoblock circumvention.
9. Preserve existing tested behavior and database state. Backwards-compatible changes with migration validation.

## Recommended next work after v0.2.0, in order

1. Test live read-only Gamma/CLOB/BTC endpoints on a real networked machine. Update parsing if market API evolves.
2. Extend existing official Gamma winner reconciliation with independent authenticated Chainlink/PolyBolt 60s TWAP capture and discrepancy alerting; never claim Gamma winner data is oracle price capture.
3. Implement real-time market book WebSocket with sequence/reconnect handling, and store reproducible L2 snapshots.
4. Extend existing safe non-executable maker quote planning with proper queue-fill simulation, inventory controls, and calibrated edge estimates; no maker fill assumptions before L2 and trade evidence.
5. Historical train/validation/test partition, walk-forward comparisons and hold-out forward paper trading; risk-of-ruin and uncertainty bounds.
6. Operational hardening: private dashboard, durable database, outage fail-safe, metrics, alerting, prolonged network simulation, CI.
7. After concrete evidence and user approval only, separately implement secure live execution.

## Quick verification

```bash
go test -race ./...
go vet ./...
go run ./cmd/bot --demo
go run ./cmd/backtest --file ./data/demo/observations.jsonl
```

Report precisely which gates pass, which are blocked, and link the exact tested revision. If nothing changed, do not claim an update.
