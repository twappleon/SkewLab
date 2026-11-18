# Five independent research-only iterations (v1.4.0)

No real orders, no paper entries, no Telegram token changes, no ledger reset. Existing Railway paper entry pause remains mandatory.

1. **Integrity:** Reject duplicate markets, mismatched policy, invalid quote/market chronology, future-resolved observations and missing outcomes. Counts are for *observed* market snapshots, never for the full opportunity set.
2. **Settlement:** Compare official Gamma-resolved outcomes vs unresolved samples, report overdue (>30m past end) and P50/P95 observed labeling delay. Missing observations are never backfilled.
3. **Paired baseline:** Compare pure momentum and aligned momentum+semivariance with the market-price baseline on exactly the same market snapshot. Require at least 100 officially published same-window, same-direction, same-price-bucket prior observations before a new observation qualifies. Count distinct UTC days; do not form a day-block lower bound unless at least 30 qualified matched results across at least 10 days. This bound is an exploratory heuristic, not a formal market independence guarantee.
4. **Cost stress:** Reprice the same resolved matched observations under increasing hypothetical taker slippage and fee reserves. No claim of actual execution or legacy early exit.
5. **Fail-closed verdict:** Show missing evidence and blockers explicitly, return only aggregate data, and NEVER authorize paper/live trading. Export offline with `go run ./cmd/study-review --file /path/to/strategy-comparison-v1.json` using a user-controlled research archive.

All diagnostics are observational *hold-to-settlement* studies. Existing legacy take-profit/stop-loss behavior requires a separate prospective execution model before any manual review of resumed **paper** entries. These experiments cannot approve live trading. Credentialed raw archives must remain private.
