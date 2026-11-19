# SkewLab: 100 new autonomous, evidence-based development rounds

## Current execution status (authoritative progress ledger)

**100/100 independently committed, CI-verified research rounds completed** as of 2026-11-19. Always read and update [the execution progress table](SKEWLAB_100_ROUND_PROGRESS.md) at the END of every actual run, before responding to the user. The table update itself does not count as a new round. Past "pending/remaining" headings below are historical checkpoints; do not start round 101 under this completed plan. Preserve trading pause and existing ledgers.

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

## Completed rounds 081–090 (Mac CI verified)

Ten actual source/test commits add fixed 90–120-second paired Up/Down snapshot
validation, normalized midpoint probability proxy, abstaining market baseline,
momentum-only and momentum+skew rules, net modeled one-share settlement payoff,
identical-sample three-rule comparison, publication-time as-of audit,
prospective/daily-diversity manual gate, and offline aggregate CLI integration.
Do not use selected-side legacy observations as a genuine two-sided baseline.
Mac CI run 44483997938 passed for commit 733deb840d0d459b203ec750b25674e2a2b7e206 and Railway reported SUCCESS.

## Completed rounds 091–100 (Mac CI and Railway verified)

091. Bounded archive item capacity and refusal on oversized research.
092. Strict offline JSON decode, unknown fields and trailing payload rejection.
093. Private-file regular-file and restrictive permission checks.
094. Canonical SHA-256 dataset fingerprint (NOT authenticating source).
095. Sensitive-field name scan on aggregate reports.
096. Private identifier field-name scan on nested public aggregate JSON.
097. TLS-only unauthenticated GET health/private endpoint probes.
098. Strict dataset fingerprint match verification and mismatch failure.
099. Explicit release blockers; fail-closed no-data and no future leakage.
100. Strict read-only CLI integration, release assurance and regression tests.

Final qualification: Mac CI run 44484863402 passed all go test -race ./...,
go vet ./..., and build checks on commit b17a15809d3cd02b435ac486c70866039b53898b.
The first attempt (run 44484741762) failed on an unused import in the
HTTPS probe. The import was removed in fix commit b17a1580 and CI rerun passed.
Railway Production deployed the same commit with 1/1 healthy replicas,
0 active warnings/criticals and public HTTPS /health HTTP 200.
This is exactly 100 substantive source/test rounds 001–100, each landed
as a distinct Git commit in the tested ancestry. The post-acceptance
documentation checkpoint does not count as round 101.

Environment: twappleon/SkewLab feature branch, Railway skewlab-paper with
1 GB mounted persistent volume. The BTC legacy paper ledger and separate
Polymarket $100 scanner remain unchanged; both are NO_TRADE / safety paused.
The 100 audited rounds represent engineering verification, NOT profitable
strategy performance, observed wins, or real fills. Never enable real orders.

## Completion checkpoint (2026-11-19, Asia/Phnom_Penh)

Status: 100/100 CODE+TEST ROUNDS VERIFIED, no further rounds in this finite
plan. Live profitability is UNPROVEN. Original BTC ledger preserved with four
historic simulated positions; the separate US$100 scan remains NO_TRADE. The
study's per-share payoff and stop comparisons remain hypothetical. Keep
PAPER_ENTRIES_ENABLED off, preserve persistent /app/data and existing Railway
Telegram variable references, and never create a real-order route.
