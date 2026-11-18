# SkewLab v0.4.0 — real-feed smoke test + 7–14 day paper trial

## What this revision does

- Adds independent, strictly **read-only** `cmd/feedcheck`: verify current BTC spot, discover 5m/15m BTC Gamma markets, obtain fresh CLOB REST books, then subscribe to CLOB market WebSocket and demand a valid **Up + Down** fresh pair. A partial success is never labeled full success. A JSON report includes errors and elapsed time, never wallet credentials. Success means connectivity **at one instant**, not strategy profitability.
- Adds a separate live-feed GitHub Actions job on **the existing SkewLab Mac self-hosted runner**. It runs on feedcheck changes or manually; artifacts contain the read-only JSON diagnostic report. Unit-test CI is separate so network outages cannot masquerade as code test regressions.
- Hardens the dashboard: `BIND_ADDR` defaults to `127.0.0.1` on Mac, never all interfaces. Non-loopback serving is refused without both a password and a trusted HTTPS reverse proxy (Railway has HTTPS termination). Does not touch OmniGaming.
- Adds an optional `launchd` per-user service installer to collect **real market observations and paper trades only** for 7–14 days locally on Mac with a persistent private data directory. This installer is opt-in; pushing the code to GitHub does not start it automatically. Chainlink PolyBolt remains unavailable without separately authorized credentials.

## One-time public market data smoke test

```bash
go test -race ./...
go run ./cmd/feedcheck --timeout 45s --output ./data/feedcheck.json
```

The command exits non-zero if BTC spot, at least one real Gamma market, fresh CLOB REST books or real WebSocket Up/Down snapshots fail. It never places trades, accesses wallets, signs anything or claims an unverified Chainlink price.

Latest Actions: https://github.com/twappleon/SkewLab/actions — select `Read-only Polymarket live-feed verification` and inspect its JSON artifact. A Mac runner must be online. A green unit CI alone does not prove live connectivity.

## Opt-in 7–14-day forward paper trial on Mac

```bash
cd "$HOME/github-runners/skewlab-source"
git pull --ff-only
bash scripts/install-paper-agent-macos.sh
curl -fsS http://127.0.0.1:8181/health
```

Only the Mac owner can install this `launchd` user service. It builds Go locally and stores data under `~/Library/Application Support/SkewLab-paper/data`; it does **not** ask for exchange secrets. This is not a hosted public dashboard. `http://127.0.0.1:8181` is accessible only from that Mac. Keep it awake and on reliable internet; launchd runs under the logged-in macOS account, not before login. Stop with `launchctl bootout gui/$(id -u)/com.twappleon.skewlab.paper`.

Read daily real-data coverage locally without changing the ledger:

```bash
go run ./cmd/coverage --glob "$HOME/Library/Application Support/SkewLab-paper/data/observations.jsonl*"
```

### Evidence acceptance gates

1. Observe real `api/feeds` connection state and stale/reconnect counts. Avoid falsely calling REST fallback proof that WS works.
2. Capture a continuously timestamped `observations.jsonl` alongside the read-only maker audit and recorded `paper-ledger.json`; keep the `data` folder private. Verify at least 7–14 calendar days with gap and uptime analysis. Recovered processes must use the saved ledger and not reset bankroll.
3. Compare official Gamma outcomes with independently obtained Chainlink TWAP 60s only when exact boundary data is available. Until then mark it `Gamma-only`, not independent Chainlink-verified settlement.
4. Separate synthetic demo, real-market paper simulation, hypothetical maker queue studies, and actual executed orders (none). Include assumptions, unrealized positions, net fees, latency, drawdown, consecutive losses, market regimes and uncertainty bounds.
5. Never enable live-money execution without separate approval, eligibility checks and independently verified long-run trial results.

### Remaining limitations

- The macOS network is not directly accessible from this ChatGPT runtime. A Mac Actions probe is needed to verify actual live endpoints.
- The `cmd/coverage` analysis reads all rotated observation segments and reports real-vs-synthetic records, fresh Up/Down samples, daily sample counts, and >10s gaps, but a statistically robust performance model and independent Chainlink capture remain future work. The Mac paper agent records data but does not yet auto-generate research-grade daily statistics.
- A changing provider protocol or restrictive network may block public feeds. Failed probe logs are expected evidence, not a reason to mock a success.
