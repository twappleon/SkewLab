# v1.5.0 — five-round performance repair (research-only)

Baseline: SkewLab v1.4.0, existing four earlier paper closures: one win, three stop-loss exits, -$38.26 realized (historically observed). These historical trades are NOT rewritten or erased. Sample count is far too small to establish predictive advantage, and the 5% stop is only a trigger when quotes and exit depth allow. Do not use this release as permission to trade.

## Five completed engineering rounds

1. **Friction and break-even hurdle:** same-snapshot ask-to-bid, two-sided taker fees, modeled extra reserve and slippage. A `cost_only_stop_breaches` observation means an immediate round-trip loss beyond the 5% stop *without a market move*. The fixed-time research archive lacks exit bid depth, so all these estimates are **non-executable diagnostics**. Median settlement break-even win probability is only a pricing hurdle, not an estimated win rate.
2. **Worst-case daily risk budget:** future paper entries, if ever authorized, now reserve the **entire potential $20 stake** per outstanding contract instead of assuming the 5% stop caps loss. Same-UTC-day realized losing trades consume the $30 risk budget, winners do not refill it. Full stake can be lost on resolution. With a $30 cap and $20 stake, at most one simultaneously outstanding full-stake position is permitted; current `PAPER_ENTRIES_ENABLED` stays off.
3. **Prospective incremental alpha:** compare the same observed 5m/15m markets where momentum+semivariance genuinely *disagrees* with the contemporaneous market-price baseline. Require >=100 earlier officially resolved same-window outcomes, >=30 later settled disagreements across >=10 UTC days. The response is a paired, modeled hold-to-expiry difference, not live fills or a proven early-exit strategy.
4. **Day concentration review:** expose the share of settled observations originating on the busiest UTC day, and count observed independent UTC dates. One hundred overlapping outcomes on one day is **not** independent evidence. Distinct days are still no guarantee of independence.
5. **Fail-closed Go/no-go:** `/api/study-review` now contains `model_version: 1.5.0`, friction, incremental alpha and day diversity plus named blockers. It stays publicly aggregate-only. The private raw research API remains authenticated, and POST is rejected. No `PAPER_ENTRIES_ENABLED=1`, no signer, no live orders.

## Verification

`go test -race ./...`, `go vet ./...`, compile the bot and study-review CLI. On Railway confirm volume, single replica, `PAPER SAFETY PAUSE` log, HTTP 200 public `/api/study-review`, HTTP 405 POST and HTTP 401 unauthenticated private export. Check exact GitHub CI and deployment commit; **never equate a green test with profitability**.

## Remaining blockers

- Need adequate real resolved fixed-time prospective samples per 5m/15m cohort, at least 10 distinct UTC days, with no look-ahead and no stale quotes.
- Need independent oracle boundary verification when available and realistic modeled execution latency and bid-side depth before any realistic early-exit evaluation.
- A positive historical mean does **not** authorize new paper entries, even with all research checks green. Explicit user approval and a dedicated paper reactivation gate would be separate work.
