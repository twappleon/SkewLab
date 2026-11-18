# SkewLab v1.3.0 — cost-aware prospective shadow study
- Remains paper-only with PAPER_ENTRIES_ENABLED disabled and emergency MIN_MOMENTUM_PCT=999 unchanged.
- Only v1.1+ fixed-time 90–120s after-open samples; do not mix the old signal-selected research archive.
- Compares momentum, momentum+semivariance, and market-price favored side against the exact same observed markets.
- Before a shadow proposal, require 100 published-before-capture outcomes in the same duration, side and 20-cent ask price bucket. Use 95% Wilson lower bound (correlation caveat) to estimate win probability; modeled cost = ask +0.002 slippage, date-aware taker fee, +25bps reserve. Require at least 2% modeled conservative ROI, bid/ask spread <=0.04, ask liquidity >=1 share.
- Each proposal is not a position or fill; score only once its official Gamma outcome is resolved. Compare to the same timestamp market-price baseline, with the same modeled cost, per one hypothetical share held to settlement.
- >=30 subsequently resolved shadow proposals per window is only an exploratory sample threshold. It never enables trading; no live wallet/order code.
- Model cannot validate the legacy early-exit execution, correlated observations, real queue fills, unobserved missing slots, independent Chainlink outcome proof, or changing market regimes.
- Verify go test -race ./..., go vet ./..., Mac runner CI, Railway health, anonymous /api/performance and authenticated /api/state /api/export.csv. Keep existing persistent volume untouched.
