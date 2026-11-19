# SkewLab: 100 new autonomous, evidence-based development rounds

Scope: `twappleon/SkewLab`, Railway `k-signal-paper-lab` service `skewlab-paper`.
This is an **incremental** plan; a round counts ONLY if its own substantive
implementation and test have landed in a Git commit, the relevant tests pass,
and a verifiable report links the commit. No empty commits or fabricated P&L.

The first ten rounds are a baseline, **not 100 completed rounds**:

1. Fail-closed pre-settlement observation and chronology validation.
2. Market-id deduplication and deterministic ordering.
3. Monitoring-window coverage with unknown-denominator handling.
4. Chronology-safe Brier/log-loss score for quote probability proxy.
5. Date-aware fees and minimum break-even win rate.
6. Wilson lower bound with at least 100 historical labels spanning ten UTC days.
7. Forward-only outcome frontier and prospective evaluation.
8. Explicit worst-case binary loss vs $2 daily stop for $100 study.
9. Paired, same-observation sampled-side filter comparison with abstentions.
10. Read-only CLI for existing comparison archive and QA observations.

## Next 90 rounds: strict continuation

11-20: capture coverage / source-of-truth timestamps / gaps / quality tests.
21-30: official resolution delay, conflicts, and independent oracle caveats.
31-40: spread, visible depth, quote age, market event/correlation audit.
41-50: historical probability calibration, Brier/log-loss confidence.
51-60: strict prequential walk-forward, market/time regime comparisons.
61-70: variable fee, reserve, stress tests, worst-case drawdown.
71-80: aligned-counterfactual exit replay with missing-data classification.
81-90: no-lookahead paired baseline comparison and uncertainty gates.
91-100: durable state/CI/real HTTPS test, README and release QA.

Each future run: discover latest branch head; read this document and commit
history; deliver up to ten entirely NEW rounds with tests. Checkpoint only
confirmed successes, report failures honestly, and do not infer profitability
from four legacy trades. Do not touch other systems' ledger or secrets.
Keep `PAPER_ENTRIES_ENABLED` off, never add a real-order path, preserve the
persistent `/app/data` volume and existing Telegram variable references.
Do not claim that a paper-only scanner has executed trades.

## Round 1-10 test summary

All ten rounds were tested individually with `go test ./internal/research100`
(or the relevant new CLI test); the release also requires `go test -race
./...`, `go vet ./...`, and at least one verified Mac CI result.

Source: `internal/research100`, `cmd/research100`. Sample input remains
private; the new CLI cannot read Railway's protected volume directly.

## Completed rounds 011–020 (Mac CI verified)

Round-specific source and tests are introduced as independent Git commits. The
GitHub branch was updated at `d13ac44db65ce56cdec315ec722df89cc1c0f123` and
Mac CI run `44480874444` passed race tests, vet, and build on the owner runner.

11. Observed quote-age P95 and as-of filtering.
12. 5m/15m capture-to-expiry horizon.
13. Candidate interarrival gaps without fabricated monitoring uptime.
14. UTC day diversity / concentration warnings.
15. Publication delay and 30-minute unresolved backlog.
16. Bid/ask spread measurement without fill assumptions.
17. Separate quote freshness and collector staleness by market window.
18. As-of chronology: future captures and future outcome labels.
19. Read-only fail-closed capture quality gate with explicit blockers.
20. Offline CLI integration and regression test for quality reporting.

The feature branch currently hosts a separate US$100 Polymarket scanner.
Never replace it with main, merge existing portfolios, expose protected raw
archives, or re-enable PAPER_ENTRIES_ENABLED. Remaining plan: 031–100 after the next 10-round CI gate.

## Pending release qualification: rounds 021–030

These ten independent Go source/test changes cover published-outcome timing,
expired market backlog, as-of outcome frontier, malformed raw settlement
chronology, matured-market labeling denominator, contradictory outcomes,
external claim disagreement without oracle attestation, history revision
checks, pending age quantiles, and combined CLI read-only review.

No source is presumed independently audited. No gain, loss or executable
order is created. This batch counts only after GitHub Mac CI and Railway
health checks on the new branch head.
