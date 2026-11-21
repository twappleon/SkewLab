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

## Telegram operational alerts
- When the authorized owner configures the existing `@crytotradebot`, send *from* the bot to the authorized numeric Chat ID; never send to the bot username itself.
- Never ask to paste tokens into chat or commit them to Git. Runtime secrets are `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` / `TELEGRAM_OWNER_ID`.
- Only send real event notifications; no synthetic, every-tick or startup messages. Throttle signal and outage alerts. Distinguish simulated positions and Gamma settlement from real fills/independent oracle proof.
- Keep alert delivery non-blocking to trading. Log sanitized failures, persist delivery IDs and do not count alerts as profitable trades. CI must use mock Telegram endpoints only.

## SkewLab execution progress ledger (binding)

For any SkewLab 100-round development/run/update task, read and update `docs/SKEWLAB_100_ROUND_PROGRESS.md` at the end of the run. Append the run's outcome, before/after count, evidence (GitHub commit and CI, Railway when relevant), blockers and next action; echo the table summary and link in the user reply. A documentation-only commit is never an additional round. The 001–100 finite plan is already completed; do not relabel unrelated work as 101. No automatic trading, no ledger reset, no credentials disclosure. If the report write fails, say so clearly.

## Mandatory behavior-driven development (BDD) for **every** round

**A Go unit test and a green CI are not by themselves BDD acceptance.** Every new development round and each historical BDD remediation must pass this sequence: (1) write/review `@round-NNN` Given/When/Then feature in `docs/bdd/`, (2) map each behavior to exact Go test function and implementation in `docs/bdd/manifest.json`, (3) revise the feature and its mapped test whenever behavior or requirements change, with incremented revision and `revision_reason`, (4) run mapped scenarios plus `go test -race ./...`, `go vet ./...`, and the `cmd/bdd-gate` CI guard, (5) accept BDD only after exact-SHA Mac CI PASS and explicit behavior/edge review, (6) update `docs/SKEWLAB_100_ROUND_PROGRESS.md` after each execution. A failed, missing, cancelled, changed-but-unreviewed scenario is BLOCKED, never PASS. If behavior is unchanged by a refactor, document the no-change rationale, rerun the mapped BDD and record the evidence.

The completed historical 001–100 engineering rounds are **not retroactively BDD accepted**: track a separate BDD 0–100 count; backfill using real Gherkin and tests, not generated PASS labels. The BDD gate blocks production Go changes lacking matching feature, manifest and mapped Go-test revisions within a pushed change set. `go run ./cmd/bdd-gate --require-all` must stay red while there is legacy BDD debt. See `docs/bdd/README.md`. No actual orders or unauthorized paper reopening.
