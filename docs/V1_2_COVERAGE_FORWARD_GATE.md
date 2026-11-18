# SkewLab v1.2.0 — Fixed-time coverage and forward-only quality gate

Read-only evidence tracking. Never enables PAPER_ENTRIES_ENABLED, changes the
paper ledger, fills, fees, strategy rules, or sends a live trading order.

## The denominator is not verified feed uptime
Each 5m and 15m series starts its theoretical coverage denominator at the
FIRST observed valid fixed-time market sample within the last 14 days. Count
only fixed-time 90–120s windows that have completed, except the first already
observed slot. Before the first valid sample, coverage is UNKNOWN (null).
Gaps may reflect Railway downtime, missing quotes, missing BTC history or
closed markets, so do not report this percentage as a provider availability SLA.
Never backfill or fabricate missing samples.

## Settlement and prospective research
For resolved/eligible percent, use only samples more than 30 minutes past
market expiry. A missing Gamma label after that point is a backlog requiring
review, not proof of provider failure or a winning/losing trade. Reject future
resolution times and invalid snapshots.

The dashboard reports the same-window 5m/15m baseline, momentum-only and
momentum-plus-semivariance forward-only outcomes. Each prospective evaluation
requires 100 previously officially published same-window results and then
at least 30 later resolved observations. Its hypothetical one-share cost and
expiry payout are NOT the legacy early-exit trade mechanics. Correlated market
outcomes and incomplete market sampling limit inference.

Public /api/performance carries only aggregate counts, quality status, and
theoretical rates; private /api/strategy-comparison continues requiring auth.
No candidate authorizes simulated or real entries automatically. Existing
volume and previous losses remain unchanged.
