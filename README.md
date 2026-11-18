# SkewLab — Polymarket BTC Quant v0.2.0

**Go paper-trading research app, not real-money trading software.** Targets Polymarket BTC 5m/15m Up/Down markets, uses Binance/Coinbase *spot* observations for 30s momentum and 60s upside/downside semivariance, fetches public Polymarket Gamma and CLOB market data, performs conservative hypothetical taker transactions, and produces **non-executable** maker quote research. Real orders, wallet authentication and real-money trades are deliberately absent.

> **Data provenance matters:** Synthetic demo data is not real performance. Spot BTC is not Chainlink TWAP. Gamma-reported official winning outcome is not an independent oracle capture. Proposed maker quotes are **never** counted as trades or profits.

## Start / verification

Requires Go 1.23+; no external Go modules. Open local dashboard at `http://localhost:8080`.

```bash
go test -race ./...
go vet ./...
go run ./cmd/bot --demo           # Clearly labeled synthetic demo
# OR:
go run ./cmd/bot                  # Real public feeds, paper execution only

# Replay your recorded (non-synthetic) observations:
go run ./cmd/backtest --file ./data/observations.jsonl

# Audit your separately captured Chainlink 60-second TWAP ticks:
go run ./cmd/oracle-audit --file ./chainlink-twap.jsonl \
  --start 2026-11-18T03:20:00Z --minutes 5 --official UP
```

**Important:** `/cmd/oracle-audit` accepts only JSON lines like `{"time":"2026-11-18T03:20:00Z","price":94201.45,"source":"chainlink-twap-60s"}`. Those numbers are illustrative schema samples, **not observed Chainlink prices**. Exact window-edge ticks and the exact source label are required; if missing, it refuses to determine a direction. The CLI is an **independent comparison tool**, not an oracle downloader. Direct authenticated Chainlink/PolyBolt streaming is **not yet implemented**.

A macOS user may double-click `start-demo-mac.command` or `start-live-paper-mac.command`. To use a different port, set `PORT=8081`.

## Strategy and safety

- Spot momentum `100 × (BTC_now/BTC_30_seconds_ago − 1)`, verified source consistency.
- Semivariance skew `(sum(positive log returns²) − sum(negative log returns²)) / sum(all squared log returns)` on a 60-second window.
- Combine thresholds: default ±0.02% momentum plus ±0.20 skew. Silence or gaps in spot feed stop new trades.
- Fetch Gamma events for current `btc-updown-5m-<epoch>` and `btc-updown-15m-<epoch>` markets and map Up/Down to token ID by outcome name. Confirm CLOB `/book` quotes at current executable best ask/bid; reject quotes with stale/missing timestamps (>12s old), crossed books, unavailable spot or insufficient displayed size.
- Simulated taker position: starting $1,000; $20 per entry; ≤$80 allocated; ≤$30 daily observed loss; ≤35 daily entries; 35s spacing; don't enter within 65s of market close. A **net ≥1%** gain after simulated fees and slippage closes the position; -5% triggers a simulated stop if executable bid depth exists. Open positions nearing expiry attempt to exit. These are assumptions, not guarantees of tradability.
- If an entry cannot exit before expiry, show `PENDING_RESOLUTION` with **zero conservative valuation** until an official published market result can be checked. Missing/ambiguous results **do not** create a win, loss payout or fictional settlement.
- Confirm a closed Polymarket Gamma market has a matching slug + condition ID + exact binary outcome prices (1,0). Reconcile with `SETTLED` **once only**. Repeated messages cannot inflate the balance. This trusts Gamma's published resolved outcome, **not** independent verification of the underlying Chainlink TWAP.
- **Maker research only:** create a buy/sell quote idea for the signal-aligned outcome using the actual CLOB tick size; reject crossed, tight or near-expiry books. There is no market-order submission, queue simulation, position inventory enforcement, or maker fill booking. A proposed sell quote may require owning shares first. Do not interpret it as realized spread or maker rebates.

## Fees — important November 2026 policy change

Official Polymarket BTC 5m/15m DeFi taker fee is modeled *per share* as:

`fee = shares × coefficient × price × (1 − price)`

- Through **2026-11-20 23:59 UTC** (exclusive): coefficient **0.07**.
- From **2026-11-20 23:59 UTC**: coefficient **0.05**.
- Maker platform fee assumed zero; maker rebates / taker rebates conservatively **not counted**.
- `Config.FeeBps` / JSON `assumed_fee_bps_per_fill` is an **additional conservative cost reserve** (new default 25bps applied to notional per fill), not the Polymarket fee itself. Old saved v0.1 ledgers that included 100bps keep that same extra reserve rather than silently rewriting historical cost assumptions.
- The default fixed price slippage remains 0.002 share price each way. Fees and assumptions can still differ from actual execution and dynamic market-level parameters.

References: [Polymarket trading fees](https://docs.polymarket.com/trading/fees), [fee & rebate changes effective Nov 20](https://docs.polymarket.com/programs/fee-and-rebate-changes-november-2026), [resolution source](https://docs.polymarket.com/concepts/resolution).

## API / Dashboard

- `/` — responsive RWD Chinese dashboard, visibly labeled `SYNTHETIC DEMO` or `LIVE DATA / PAPER TRADING`.
- `/health` — liveness, **not** proof that market feeds are healthy.
- `/api/state` — snapshot including fee coefficient and **non-executable maker_quotes**.
- `/api/export.csv` — taker paper positions and official-result reconciliations.
- `/api/observations.jsonl` — raw observations and sampled public market quotes.

Persistent files reside under `./data` (or `--data-dir`), and demos always use `./data/demo`; avoid committing them to Git. Paper snapshots are written via atomic rename. Observations append to JSONL. A `PENDING_RESOLUTION` position generated in **v0.1** lacks a condition ID, so v0.2 will **not auto-settle it**; it must be reconciled with separately verified source identifiers rather than trusting a guess.

## Deployment and remaining work

Run the Go service on an always-on container (Dockerfile/Railway provided) with a **persistent disk/volume** and **HTTPS access control**. For a publicly reachable Railway deployment set `DASHBOARD_USER` and `DASHBOARD_PASSWORD` via platform secret environment variables **before launch**; otherwise Railway startup refuses to serve a live paper dashboard. Protect raw observations and portfolio history. This build has **not** been verified against the real public market endpoints because the local build environment has no outbound DNS. Do not publish live earnings or connect a wallet without additional tests.

Still open:

1. Validate actual Gamma/CLOB/BTC schema and access from a networked environment; test market switching, API downtime and reconnect. The late-November **PolyBolt Chainlink API key / provider requirement** also needs implementation and authorized credentials to independently capture Chainlink 60s TWAP readings.
2. Add reliable market WebSockets, orderbook event sequencing, queue-aware maker fill simulation, maker inventory controls, and realized fill attribution; only then measure maker performance.
3. Replace JSONL with durable database, log rotation and alerting; test crash/restart and 7–14 days of real-market paper observations.
4. Walk-forward backtest with realistic market outcomes, sharp-cost fee schedule, slippage by available depth and out-of-sample validation. No real-money trading until explicit human approval, eligibility verification and robust safeguards.

**Live network status: not verified. Cloud deployment: not performed. True profitability: unknown.**

## QA Screenshots (synthetic demo only)

Synthetic dashboard screenshots are included in the downloadable v0.2.0 release ZIP from the development chat; image files are not tracked in this source repository. The screenshots do not show real trading results.
