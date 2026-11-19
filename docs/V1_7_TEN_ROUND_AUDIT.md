# SkewLab v1.7.0 — Ten rounds of replay credibility repair

This release is **not** a profitable-strategy claim. It leaves the existing
`PAPER_ENTRIES_ENABLED` safety pause and all real trading disabled.
No balances, historical P&L or stored paper positions are rewritten.

1. Reject malformed and future timestamped bid/ask quotes.
2. Refuse to use quote information that postdates the saved observation frame.
3. Only replay exits strictly after entry and before market expiry.
4. Reject duplicate market identities within one archived frame.
5. Report hypothetical normal, doubled slippage and quadrupled reserve friction.
6. Count fresh observed books without extrapolating feed uptime.
7. Audit ledger reported cash and realized P&L against all recorded positions.
8. Report paired alternative-minus-legacy P&L only when both are observable,
   and require at least 30 pairs across 10 UTC days for each market duration.
9. Return a fail-closed research evidence gate with specific blockers and
   **trading_authorized=false** unconditionally.
10. Expose that gate in existing read-only replay CLI and public aggregate
    review route. CI is run on the authorized owner Mac, for vetted branch
    pushes only (not fork pull requests).

The report is observation-limited: empty or incomplete archives are NOT a
zero-loss backtest. Quote depth, timestamp validity and cost assumptions do
not prove executable real fills. A larger stop in the alternate model does not
justify changing the running strategy. All paper and real entries stay paused.
