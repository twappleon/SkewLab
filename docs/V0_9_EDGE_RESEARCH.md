# v0.9.0 — Time-causal edge research; no trading authorization

## Why the safety pause remains

The previous four completed paper positions produced -US$38.26, three recorded as stop losses. The sample cannot establish predictive edge. The observed immediate round-trip taker friction at approximately 50-cent prices already exceeded the prior 5% stop trigger. **PAPER_ENTRIES_ENABLED is still disabled unless explicitly changed by the owner**, and the separate emergency `MIN_MOMENTUM_PCT=999` remains in Railway.

## How samples are collected

- Candidate observations are recorded **once per active market** even with new paper entries paused, using the original reference 30-second momentum (0.02%) and 60-second semivariance skew (0.20). Those research thresholds are not used to grant trading permission.
- Record one timestamped market identity, condition ID, signal side, market window, bid/ask and depth, market end, source quote timestamp, momentum and skew. Quote must be complete and at most 12 seconds old. A sample is **not** a hypothetical or actual order.
- Only after market expiry, a background task asks the official Polymarket Gamma `Settlement` endpoint for the *same* market slug and condition ID. No match, incomplete outcome, or unavailable network => remains pending, never infer from the BTC spot price.
- The durable `edge-research.json` is written atomically under `/app/data`. Samples survive restarts and are not overwritten on corrupted journal, oversized history or disk failures. Outcomes are labeled only after they are published. Journal supports up to 50,000 samples, after which research halts safely until archival is explicitly planned.
- Gamma resolution is **not** independent Chainlink TWAP proof; this distinction must remain visible in output.

## Research-only expected-value model

The offline `cmd/edgecheck` estimates the proportion of correct calls within the same market duration (5m vs 15m), side (UP/DOWN) and 20-cent ask-price bucket. It requires at least **100 independently resolved market outcomes** in a cohort. It applies a one-sided 95% Wilson lower bound and checks whether the projected settle-to-expiry payoff at that bound exceeds modeled entry cost by at least 2%. No future-labeled samples can be used to decide a past candidate.

```
go run ./cmd/edgecheck --file ./data/edge-research.json
```

This is **not** a forecast calibration guarantee. Markets within adjacent minutes are correlated and stationarity cannot be assumed. The model assumes holding through market settlement; the existing paper-trading strategy exits early. Real-world taker fees, slippage, delayed/failed orders, missing public events, and prediction-market pricing may make the simple projection optimistic. No matter what it returns, `RESEARCH_CANDIDATE_ONLY` does not activate trading.

## Observability and access controls

- Passwordless `/api/performance` exposes only aggregate research candidate/resolved/pending counts and storage warning, not condition IDs, slugs or raw data.
- Detailed `/api/edge-research` requires the existing Dashboard password and returns a bounded recent subset plus an offline-style walk-forward summary; it is not a trading endpoint.
- The public homepage has an explicitly research-only status section, and the older private raw trade/quote/CSV APIs remain protected.
- `go test -race ./...` validates no synthetic outcome, outcome chronology, deduplication after restart, corrupted-file fail-closed, price-bucket separation, and look-ahead protection.

## Follow-up before considering any new paper entries

1. Gather 7–14 days of independent market candidates and official results; verify missing data rate and source integrity.
2. Compare a no-signal baseline and pre-specified 5m/15m cohorts using strictly forward-time labels, then evaluate uncertainty and liquidity costs.
3. Build a separate prospectively tested, time-aligned risk model for **actual** legacy early-exit behavior. A positive hold-to-settlement Wilson estimate is **not** sufficient for reopening paper entry.
4. Obtain explicit owner approval before enabling any paper entry flag. Real-money order routes do not exist.
