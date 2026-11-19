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
archives, or re-enable PAPER_ENTRIES_ENABLED. Remaining plan: 051–100 after the next 10-round CI gate.

## Completed rounds 021–030 (Mac CI verified)

These ten independent Go source/test changes cover published-outcome timing,
expired market backlog, as-of outcome frontier, malformed raw settlement
chronology, matured-market labeling denominator, contradictory outcomes,
external claim disagreement without oracle attestation, history revision
checks, pending age quantiles, and combined CLI read-only review.

No source is presumed independently audited. No gain, loss or executable
order is created. GitHub Mac CI run `44481364383` passed on commit
`65539cf91580d58480a4bea60df0382b05b6051b`; Railway deployment
reported SUCCESS and a running replica.

## Completed rounds 031–040 (Mac CI verified)

10 independent code/test rounds introduce bid+ask depth floors, hypothetical
liquidation capacity, top-book depth imbalance, ask-price bucket distribution,
malformed raw quote rejection, same-book model friction, joint age/spread flags,
5m/15m temporal overlap, layered quote filters, and read-only CLI integration.
None of these simulate a real fill or authorize trading. Mac CI run 44481856140 passed on commit
83ccc5811ba19d2d643ff318eff20e35714ddec4, and the Railway deployment
reached SUCCESS with one running replica.

## Completed rounds 041–050 and 051–060 (Mac CI verified)

Ten independent code+test rounds introduce 5m/15m quote-proxy Brier/log-loss,
price reliability bins, explicit probability clipping, a constant 50% Brier
baseline, per-window/side Wilson intervals, same-observation paired Brier,
price-bound sensitivity, chronological score drift, conservative calibration
gates, and offline CLI review integration. This is in-sample research using
selected side observations, NOT a calibrated or executable trade model.

CI on the final Git ancestor and a healthy Railway deployment are required
before counting these rounds as completed. No trading entry flag change.

## Pending release qualification: rounds 051–060

10 distinct source/test commits add strict time-split train/holdout membership,
as-of prior labels, held-out day diversity, UTC hour and weekday/weekend regime
scores, momentum strength groupings, semivariance alignment, abstention-aware
comparison, a forward-only evidence gate, and an offline CLI research report.

The first CI attempt on rounds 041–050 failed on a mistaken Brier regression
test expectation (run 44482306473). Round 056 corrects the test by supplying
a genuinely more confident winning quote. This is an explicit test correction,
not a fabricated market result or change in profit. The corrected full Mac CI run 44482771167 passed on commit
33112121b2e52f5084f61fea651bd6d1021d47b8. The prior failed run
44482306473 is retained as diagnostic history, not represented as success.
The code NEVER enables any paper or real trading route. Remaining: 091–100 after the paired comparison batch passes CI.

## Completed rounds 061–070 (Mac CI verified)

Each round contains a distinct source and test: date-aware fee transition,
three slippage levels, three reserve assumptions, same-book round-trip
friction, worst-case US$100 stake-at-risk, modeled stop headroom, required
settlement win rate, one-share settlement counterfactual, hypothetical UTC
loss-day analysis, and an explicitly non-executable cost review in the CLI.
Never turn this cost modeling into paper/live orders. Mac CI run 44483160891 passed for commit c50879a635de1cf755223c477f9d96e76095de80; Railway reported SUCCESS. They remain model-only research.

## Completed rounds 071–080 (Mac CI verified)

Ten distinct source/test commits implement entry evidence checks, explicit
post-entry bid witness validation, unique pairing without cherry-picking,
modeled round-trip P&L, four-second take-profit checks, 5% stop triggers,
32-second expiry proximity, same-witness 5% vs 10% threshold comparison,
missing-evidence coverage, and an optional private --exit-witness-file CLI
research report. Missing witness bytes always produce NO_VALID_POST_ENTRY_BOOK_EVIDENCE.
No public market IDs, actual order fills or profit claims are generated.
Mac CI run 44483609962 passed on commit 08b4cb85c3f72e3d8f95aea45cb527aab1b033e9; Railway reported SUCCESS with the persistent Volume retained.

## Pending release qualification: rounds 081–090

Ten actual source/test commits add fixed 90–120-second paired Up/Down snapshot
validation, normalized midpoint probability proxy, abstaining market baseline,
momentum-only and momentum+skew rules, net modeled one-share settlement payoff,
identical-sample three-rule comparison, publication-time as-of audit,
prospective/daily-diversity manual gate, and offline aggregate CLI integration.
Do not use selected-side legacy observations as a genuine two-sided baseline.
This batch is counted only after owner Mac CI and Railway pass.
