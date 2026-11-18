# SkewLab v1.6.0 — recorded-quote counterfactual replay (read-only)

## Objective
The visible portfolio's historical losses do not disappear when new paper entries are paused. After v1.5, no new model trades are opened unless the owner explicitly re-enables them. The replay is intended to *test explanations*, not to make a poor strategy look profitable.

## Five independent development rounds
1. Reject absent/malformed JSONL and duplicate observation timestamps without inventing a missing book or settlement.
2. Audit each saved entry against a PRE-entry bid/ask observation and model immediate round-trip loss, including the dated platform fee, extra reserve and both assumed slippages. Never call this an executable fill.
3. Replay the ORIGINAL 1% net take-profit / 5% net stop versus the pre-registered research-only 1% net take-profit / 10% net stop. Require correct market/condition ID, a timestamp-fresh book, and enough bid depth. If no complete exit evidence, return NO_VERIFIABLE_EXIT, not an invented P&L.
4. Compare paired results only for the same previously recorded positions, distinguish 5m/15m, require >=30 matched exits on >=10 independent UTC days before research review. Do not overwrite actual historical ledger P&L.
5. Deliver a read-only local CLI and on-demand, cached public aggregate GET /api/replay-review. Raw observations and ledger export remain authenticated. No POST or state mutation.

## Offline audit against a private Railway volume copy

```
go run ./cmd/replay-audit --ledger /path/to/private/paper-ledger.json --glob '/path/to/private/observations.jsonl*'
```

The CLI reads only local files and never sends network requests. It outputs aggregate market groups, observed entry coverage, cost-only stop risks and paired modeled exit results. Missing files/invalid records make it fail with a nonzero exit code. With 4 trades, no conclusion about real expected value is possible.

## Live public summary

`GET /api/replay-review` is a read-only, cached 3-minute aggregate. The page has an explicit *click to run* control, not a 15-second automatic reread of a potentially large JSONL archive. Each request allows up to 80MiB of source observations and 120,000 frames and reports unavailable/incomplete data rather than silently pretending 0 profit. It does not expose market IDs, tokens, strategy signals, passwords, raw position IDs or filesystem paths. Raw archives are still authenticated.

## Scientific limitations

- Exiting at a quoted best bid is a **modeled assumption**; top-of-book depth and timestamp do not guarantee execution, and a real order could slip or fail.
- Snapshots are typically spaced seconds apart. Intratick price changes and exact order queue positions cannot be reconstructed; sparse archives must be flagged.
- The alternative 10% stop may increase risk significantly. It is **not recommended or enabled**. Original stop and net profit settings are unchanged.
- Official settlements are not inferred. The replay cannot infer a market's final winner from the BTC spot price, and it cannot claim that CLOSED trades were held to settlement.
- Archive segments from older periods may be missing, rotated, or truncated. Coverage reflects only provided source bytes.
- Both paper and real trading remain disabled until independent prospective tests, live execution feasibility checks, and explicit human approval.
