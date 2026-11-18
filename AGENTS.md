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

## Recommended next work after v0.3.0, in order

1. Deploy code to a newly owned, access-protected repository and single-process Railway service with a persistent volume. No public link before a tested healthy service is reachable.
2. Validate live read-only Gamma/CLOB/BTC schemas and the public WebSocket in a networked environment. Test 5m/15m market rollovers, intermittent DNS, silent streams, forced reconnect, and crossed books.
3. Configure authorized CLOB API credentials + Chainlink API key via secrets manager. Check live PolyBolt authenticated `chainlink` TWAP 60s against official Gamma outcomes; log all mismatches without auto-overriding resolutions.
4. Gather 7–14 days of reliable timestamped real market observations with a documented uptime/error report; compare indicative maker queue estimates to genuine execution evidence before booking any fills.
5. Add full maker inventory constraints, fee changes, walk-forward strategy comparison, calibration, latency, adverse selection and uncertainty bounds; no profits from synthetic assumptions.
6. Add PostgreSQL/SQLite WAL backup and migrations, alerts, server resource metrics and recovery drills.
7. Implement secure live execution only after evidence, explicit informed human approval and all platform/jurisdiction checks.

## Quick verification

```bash
go test -race ./...
go vet ./...
go run ./cmd/bot --demo
go run ./cmd/backtest --file ./data/demo/observations.jsonl
```

Report precisely which gates pass, which are blocked, and link the exact tested revision. If nothing changed, do not claim an update.
