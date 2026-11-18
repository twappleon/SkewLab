# SkewLab v0.6.0 — 5m vs 15m paper performance

This release is read-only analytics. Underlying paper order, fee, risk, settlement, ledger persistence and Telegram credentials are unchanged.

- Public `/api/performance.market_windows`: three rows in order `5m`, `15m`, `other`. Unknown slugs are not silently discarded.
- Each row contains latest 14 local days (including today), closed trades, wins, net, gross profit, gross loss, average net, profit factor (null if no losses), win rate (null for no closes), and aggregate exit reason counts.
- Open inventory, pending resolution, future timestamps and non-finite P&L are excluded. Profit factor is undefined with zero gross losses. Settlement is classified separately from pre-expiry stop/take-profit outcomes.
- No individual trade IDs, market slugs, token addresses, private positions or raw observations appear in the public response. It **does** disclose coarse strategy performance publicly.
- With fewer than 30 closes per market, label results "insufficient"; at 30–99, "preliminary"; at 100+ these are research samples, not proof of profitability.
- The daily Telegram summary includes per-market 14-day net P&L and exit categories, but only if Telegram is already configured; deployment does not add a new Telegram token or automatic smoke-test send.
- Validate using `go test -race ./...`, `go vet ./...`, and Go builds. Check Railway `/api/performance` without login and `/api/export.csv` still HTTP 401 without password.

Known limitation: the daily-loss cutoff prevents new entries after a threshold is observed; stop/expiry can produce larger realized losses than that threshold. The ledger has no independent statement of actual Polymarket fills because all trades are simulated.
