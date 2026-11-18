# Safety update v0.7.1 — stop new paper entries pending positive-edge validation

On 2026-11-18, the first four completed paper positions lost $38.26, with 3 simulated stop-loss closes exceeding the nominal 5% trigger. A no-movement 50c/50c contract round-trip loses ~7.97% under the app's fee/slippage buffer (0.07 per-share taker coefficient, 0.002 simulated slippage per side, 25 bps buffer per fill). This alone can trigger the -5% stop within four seconds; the momentum+semivariance signal cannot be treated as a calibrated probability estimate.

New *live-data paper entries* are disabled by default and require deliberate `PAPER_ENTRIES_ENABLED=1`. Existing paper exits, expiry reconciliation, monitoring, and Telegram reports continue. The engine also rejects entries whose immediate cost-aware round-trip loss equals or exceeds the configured stop-loss threshold, or whose visible bid-side depth cannot support hypothetical exit size. This is only a necessary rejection test; it does not prove positive expected value. Historical ledger losses are not altered.

The Railway emergency setting `MIN_MOMENTUM_PCT=999` remains a second independent pause until the strategy is reviewed. **Do not re-enable entries yet.** There are no real orders or wallet credentials.

# SkewLab — Polymarket BTC Quant v0.7.0

## Market-window performance diagnosis (v0.6.0)

The passwordless public aggregate endpoint now includes `market_windows` with 5m, 15m and unknown classifications for the latest **14 calendar days in Asia/Phnom_Penh**. Groups show completed paper positions, realized P&L, win rate, average, profit factor where defined, and aggregated exit reasons (take profit, stop loss, pre-settlement exit, official settlement, other). No market slug, position ID, user ID, token ID or individual trade is disclosed. The public landing page displays this breakdown; Telegram daily paper messages include the 5m and 15m totals. This is paper-only analysis, not a claim of live profits.

**Risk note:** The observed 2026-11-18 Dashboard snapshot had 4 paper closes, -$38.26 realized P&L, 25% win rate and PF 0.02. This is too small a sample to conclude a stable edge. The daily $30 entry cutoff is not a guaranteed maximum loss; expiry or unavailable bids may cause overshoot. Do not enable real orders or increase notional based on early results.

## Seven / fourteen day performance study (v0.5.0)

The passwordless aggregate dashboard at `/` now shows 7-day realized P&L, 14-day average net result per closed paper transaction, a profit factor where defined, and an explicit local-calendar-day table (Asia/Phnom_Penh). The same read-only aggregates are added as `analytics` in `/api/performance`. Daily Telegram paper summaries now include 7/14-day results. **No new trading permissions, execution endpoints, or real orders were added.**

Only `CLOSED` and `SETTLED` positions with actual exit timestamps enter these calculations. Open inventory, unresolved settlement, future-dated closes and zero sample counts never fabricate realized profits. Profit factor is `null` when there is no measured gross loss (undefined), and average net result and win rate are `null` for empty samples. All figures depend on configured fee/slippage assumptions; realized **paper** profit is not actual transferable cash.

Aggregates expose no trade IDs, token IDs, or market identifiers. `/admin`, `/api/state`, `/api/export.csv`, and raw archives remain protected by `DASHBOARD_PASSWORD`. If ledger aggregate close count disagrees with the recorded completed positions, the public page warns about incomplete history instead of asserting trustworthy statistical results.

For evidence quality, fewer than 30 closed trades in 14 days is marked insufficient, 30–99 preliminary, 100+ a research sample (still not predictive or proof of future profits). Read [performance methodology](docs/V0_5_PERFORMANCE.md).



**Status: locally tested, paper-only. No real orders, no wallets, no validated live-market profitability.** The project extends v0.2.0: 30-second BTC spot momentum, 60-second up/down semivariance, 5/15-minute Polymarket market discovery, conservative taker paper execution, official Gamma result reconciliation, date-aware taker fees, and authenticated RWD monitoring. v0.3.0 adds public CLOB WebSocket ingestion, *optional* authenticated PolyBolt Chainlink 60s TWAP capture, independent disagreement interlock, passive maker-queue research and crash-aware persistence.

## New in v0.4.0: first live-market evidence, safely

- Standalone `go run ./cmd/feedcheck --timeout 45s` tests real public BTC/Gamma/CLOB REST and **CLOB WebSocket** freshness without a wallet or orders. The command returns failure unless a real Up+Down market pair is observed. A separate Mac GitHub Actions job produces downloadable diagnostic JSON artifacts; no fabricated live successes.
- Local Dashboard now binds to `127.0.0.1:8080` by default, protecting paper trading records from unintended LAN access. A non-loopback listener requires authenticated access and HTTPS reverse proxy termination (`TRUSTED_REVERSE_PROXY=1` or Railway). `BIND_ADDR` can override deliberately.
- Optional Mac user launchd service at `scripts/install-paper-agent-macos.sh` runs the real-public-data **paper-only** bot on `127.0.0.1:8181` and persists to `~/Library/Application Support/SkewLab-paper/data` when explicitly installed on the owner's Mac. Does not modify existing OmniGaming runner.

Local observation gap analysis: `go run ./cmd/coverage --glob "./data/observations.jsonl*"` (not a profit report).

See [v0.4 live validation protocol](docs/V0_4_LIVE_VALIDATION.md).

## Run locally

Requires Go 1.23+ (standard library only).

```bash
go test -race ./...
go vet ./...
go run ./cmd/bot --demo                    # SYNTHETIC data and trades only
go run ./cmd/bot --data-dir ./data         # public market APIs; paper trades only
```

Open `http://localhost:8080`. The Mac scripts `start-demo-mac.command` and `start-live-paper-mac.command` are included. Set `PORT=8081` for a different port. Use `DASHBOARD_USER` and `DASHBOARD_PASSWORD` (environment variables) for HTTP Basic Auth, and **always use HTTPS behind a trusted reverse proxy** if externally accessible. Production Railway startup refuses to serve the live-paper dashboard without `DASHBOARD_PASSWORD`. `/health` reports liveness only, NOT live feed readiness. `/api/feeds` reports the actual stream status.

### Optional Chainlink TWAP streaming (read-only)

Reference prices were moved to the authenticated **PolyBolt** protocol. Configure the following secrets on the host using your secrets manager, **not source control, CLI arguments, or chat**:

```
POLYMARKET_API_KEY
POLYMARKET_API_SECRET
POLYMARKET_API_PASSPHRASE
CHAINLINK_API_KEY
```

No wallet/private key is used. If one of these is missing, the Chainlink data feed explicitly shows `NOT CONFIGURED` and **never invents TWAP values**. Connection to `wss://ws-live-v2.polymarket.com/ws` performs `auth` with Chainlink key on the first frame, waits for `authed`, and subscribes to `price.crypto.twap` with `{"symbol":"btcusd","window_seconds":60,"provider":"chainlink"}`. It verifies a Chainlink provider acknowledgement, BTC/USD symbol, window size, event sequence and drop count; stale/future updates and sequence gaps cause a reconnect. Captured observations are recorded in `chainlink-twap-60s.jsonl`, with the known exact-source label. Some market bounds may not appear exactly even when data is streaming; the audit requires exact timestamps and refuses to guess them.

Official protocol: https://docs.polymarket.com/api-reference/live-data/overview

### Public market WebSocket

- Read-only `wss://ws-subscriptions-clob.polymarket.com/ws/market`. After discovering active token IDs via Gamma, subscribe to 5m/15m Up and Down tokens.
- No API key is needed for public CLOB market streams. Reconnect with backoff, resubscribe on market rotation, discard old books and reject out-of-order, stale, crossed or missing books. Support both `book` snapshots and `price_change` updates; use `last_trade_price` for the non-executable queue study.
- The trading strategy only accepts a fresh **pair** of WS quotes; otherwise it fetches and timestamp-validates both CLOB REST books. WS reconnects or a stale snapshot **never authorize fictional fills**.
- Spot BTC still comes from Binance/Coinbase and is NOT the Chainlink market resolution feed. Momentum history resets when the spot venue changes. A valid signal is not sufficient to establish positive expected value.

Official protocol: https://docs.polymarket.com/market-data/realtime-data

### Passive maker queue research — never a booked trade

An indicative maker proposal now also starts a *different*, conservative research study at the observed public best bid. It stores the initial visible queue depth and hypothetical share quantity. Only later public `last_trade_price` events labeled SELL at **the exact price** can reduce observed queue ahead. Obvious duplicate timestamps are conservatively ignored; quote cancellation and stale books do not count as traded volume. Reaching the estimated queue threshold yields **`HYPOTHETICAL_QUEUE_CLEARED_NOT_ACTUAL_FILL`**, never a realized fill, inventory, rebate, PnL or booked maker strategy win. This remains an uncalibrated estimate: the hypothetical order was never submitted, and the public trade channel cannot establish the order's actual queue rank or execution. Sell-side research does not imply free inventory.

### Official resolution and independent TWAP conflicts

Gamma must publish a closed outcome with matching condition ID and exact binary winner prices. Settlement is idempotent. When both exact market boundary TWAP values have been recorded through Chainlink, the engine additionally compares `close >= open` and **blocks the settlement plus future paper entries on disagreement** (manual review required). When exact Chainlink edges are missing, Gamma-only settlement remains possible and is labelled **not independently verified**. Missing TWAP samples never become synthetic Chainlink prices. Exchange spot cannot be used for a settlement comparison.

The official 5/15-minute BTC market's Chainlink **TWAP 60s** is the settlement reference. Taker fee uses `shares × coefficient × price × (1 − price)` with the official 0.07 (before 2026-11-20 23:59 UTC) and 0.05 (on/after) coefficient. An additional conservative 25-bps notional fee and 0.002/share slippage per fill are included. No rebates are assumed. Exits require **net profit ≥1%** after modeled costs, or a risk exit, and depth checks are mandatory. Default starting balance $1,000, $20 entry, $80 exposure and -$30 observed daily loss limit.

Fees: https://docs.polymarket.com/trading/fees and https://docs.polymarket.com/programs/fee-and-rebate-changes-november-2026

## Persistence, restart and audit

One process per persistent volume. `--data-dir ./data` holds private:

- `paper-ledger.json` (atomic write + file fsync + rename + directory fsync). Corrupted/unreadable ledger **halts all paper execution and prevents overwrite**, instead of silently resetting account equity.
- `observations.jsonl` — replayable market/spot observations, 50MiB rotated segments.
- `chainlink-twap-60s.jsonl` — genuine read-only authenticated samples if connected, 25MiB rotated segments; unrotated tail is restored into the on-process audit cache after restart, but NEVER treated as a live connection.
- `maker-queue-audit.jsonl` — explicitly hypothetical study events, 25MiB rotated segments.

Rotated segment filenames carry a UTC timestamp suffix. JSONL writes are appended and synced before return. If file writes fail, **ledger persistence errors halt paper trading**. Place `data/` on a real persistent Railway volume, never an ephemeral container filesystem; configure auth before public deployment. `Dockerfile` and `railway.toml` are included but a cloud deployment has **not** been achieved or verified.

API routes: `/`, `/health`, `/api/state`, `/api/feeds`, `/api/export.csv`, `/api/observations.jsonl`, `/api/chainlink-twap.jsonl` (raw exports require dashboard authentication).

## Tests and caveats

```bash
go test -race ./...
go vet ./...
go build ./cmd/bot
go run ./cmd/backtest --file ./data/observations.jsonl
go run ./cmd/oracle-audit --file ./data/chainlink-twap-60s.jsonl \
  --start 2026-11-18T03:20:00Z --minutes 5 --official UP
```

Tests cover masked websocket handshake and PING/PONG, canonical/legacy book events, price changes, stale/out-of-order data, market rotation, trade queue, invalid provider, PolyBolt sequence gaps, ledger durability, corrupted ledger fail-closed, restart recovery, official reconciliation idempotency, and conflicting Chainlink/Gamma outcomes. Browser UI tested at 1440px desktop and 390px mobile widths using locally served **synthetic** API data; no JavaScript errors or horizontal overflow. Demo does **not** prove profitability.

### Not yet proven / next verification

1. **Production market connectivity unverified:** sandbox DNS cannot reach Gamma, CLOB, Binance, Coinbase or PolyBolt. No actual continuous production stream was obtained in this session. The binary requires a networked machine and real credentials for Chainlink. The changing 2026 Polymarket protocol may require further implementation adjustments.
2. **No cloud deployment:** this v0.3.0 revision has not yet been confirmed on GitHub or Railway/Vercel service deployed. Thus there is NO validated public URL.
3. **No real maker fills or market-maker profit:** queue research is not a true posted-order execution model. Precise order queue position, cancellation priority, adverse selection, inventory, API latency and maker rebates remain to validate.
4. **Not a 7–14 day real-market paper experiment:** no statistically supported win rate/expectancy or verified profitability. No wallets, real orders or live trading have been implemented.
5. For reliable multi-instance or long-term storage, upgrade the single-process JSONL/atomic-file store to managed Postgres/SQLite WAL with backup and migrations, add network metrics, and conduct a 7–14 day forward paper trial before discussing money.

Synthetic screenshots in `docs/dashboard-v0.3.0-synthetic.png` and `docs/mobile-v0.3.0-synthetic.png`. Never label them live trading or earnings proof.

## Telegram alerts — @crytotradebot (paper-only)

SkewLab can **send from the already owned** `@crytotradebot` Telegram bot to its authorized owner chat, not *to* the bot username. The recipient must first open the bot and press Start; for private chats use the **numeric Chat ID**, not the private account's @username. Never use `@crytotradebot` as `TELEGRAM_CHAT_ID`.

Set the following environment variables in the SkewLab Railway **service variables**, not the repository, GitHub Actions or chat:

- `TELEGRAM_BOT_TOKEN` — reuse the existing `@crytotradebot` bot token securely; obtain it from your secrets manager / BotFather if necessary, never paste it into a conversation.
- `TELEGRAM_CHAT_ID` — authorized owner or private group numeric chat ID; if absent, `TELEGRAM_OWNER_ID` is supported for compatibility with the older trading bot.
- `TELEGRAM_TARGET_USERNAME` — optional fallback for a **public channel/supergroup** where the bot has appropriate permissions; this does not reliably work with ordinary private user handles.

No token + numeric Chat ID means Telegram remains disabled, **but paper trading continues**. Synthetic demo and CI never dispatch Telegram messages. If the bot has not been started by its target user, Telegram may return HTTP 403; the status in `/api/state` (`telegram_notifications`) and runtime logs expose only sanitized error codes, never credentials.

Messages are sent only for significant events: new aligned momentum/semivariance direction (throttled 5 minutes), hypothetical paper entry/exit and net +1% take-profit or stop-loss, pending settlement and Gamma settlement, sustained public-feed outage and recovery, daily loss cutoff, oracle conflict, disk failure, and a daily 09:00 (Phnom Penh) paper account summary. **No per-tick notifications, automatic startup tests or real orders.** Successful sends are journaled as `telegram-delivered.jsonl` on the persistent volume; commit IDs based on position ID reduce duplicate alerts after restarts. Delivery is best effort: if Telegram is unavailable, the bounded queue protects trading, and individual event messages may be lost. Keep only one Railway replica to avoid duplicate strategy runs.

To verify the recipient manually **after configuring secrets**:

```bash
go run ./cmd/telegram-test                     # no message sent
# Only when ready to send one explicit test message:
go run ./cmd/telegram-test --send
```

Telegram setup is **code-complete only** until the SkewLab Railway service exists and secrets are configured there. Do not claim messages were delivered merely because CI passed.
