# SkewLab v1.0.0 — Research evidence quality gate

Read-only study, NOT an execution system. Current PAPER_ENTRIES_ENABLED remains disabled and the emergency MIN_MOMENTUM_PCT=999 must not be removed.

## New reporting

`edge_quality` in public `/api/performance` provides aggregate-only counts (no trade IDs, condition IDs, market slugs, or raw books): observed candidate total, official Gamma-labeled total, pending, >30-minute post-expiry backlog, invalid chronology, 14-day captures/resolutions, latest timestamps, 5m/15m totals, and largest matching cohort of duration/side/20-cent ask bucket.

The private `/api/edge-research` includes the same quality report beside individual research records and walk-forward output. Authentication remains mandatory.

## Research readiness semantics

- `NO_CANDIDATES_YET`: no qualifying observations have been recorded.
- `COLLECTING_INSUFFICIENT_EVIDENCE`: all cohorts have fewer than 100 valid resolved observations.
- `SETTLEMENT_BACKLOG_REVIEW`: at least one recorded candidate remains unlabeled 30+ minutes after expiry. This flags a possible backlog, **not proof** of an API failure.
- `NO_RECENT_QUALIFYING_SIGNALS`: no valid qualifying candidate captured in 24 hours; not necessarily a feed outage.
- `INVALID_RESEARCH_RECORDS` / `RESEARCH_ARCHIVE_ERROR`: stop trusting the archive and inspect storage/chronology.
- `RESEARCH_REVIEW_ONLY`: at least one cohort reached 100 observations. This is a manual review trigger, **not** authority to reopen simulated or real-money trading.

The number of recorded qualifying signals cannot be converted to overall market-data uptime or exposure frequency without an independent full-market coverage journal. Market outcomes come from Gamma and have not independently proven Chainlink TWAP conformity.

Tests: `go test -race ./...`, `go vet ./...`, `go build ./cmd/bot`. Public API must not reveal private market identity; CSV/raw research are still protected by HTTP Basic Auth. No wallet or real-money order path exists.
