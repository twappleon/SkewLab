# SkewLab v0.7.0 — model-loss diagnostics

This release explains **simulated** completed-trade outcomes. It does **not** change the strategy or identify causes unsupported by available event history.

## Definitions

- `stop_threshold_breaches`: completed positions with `net stop loss` note whose final simulated net loss is worse than configured stop (default 5%). It does not prove a market execution failure.
- `excess_stop_loss_usd`: `max(0, -realized_pnl_usd - entry_cost_usd * stop_pct/100)` summed over stop-labeled trades.
- `contract_price_pnl_usd`: `qty * (stored_exit_price - stored_entry_price)`; assumptions for fixed per-share slippage are already embedded in prices.
- `modeled_cost_reserve_usd`: `contract_price_pnl_usd - realized_pnl_usd`. This includes modeled trade fee/buffer reserves, not separately the market bid/ask spread, real latency or order execution.
- `attribution_complete`: false if historical records lack valid finite quantity/prices/costs or if accounting identity is inconsistent. These trades are excluded from cost attribution but not the net-loss category.
- `by_reason_market`: grouped 5m/15m/other and stop/take-profit/settlement/before-expiry outcomes.

## Security

Public performance exposes aggregate counts and cost estimates only. It never reveals trade IDs, market slugs, token IDs, individual timestamps or account secrets. Authenticated `/api/diagnostics` returns up to 100 normalized per-trade research records without market identifiers. CSV and administrative views remain guarded by Railway `DASHBOARD_PASSWORD`.

## Limits

Overshoots require further investigation of time, market movement, and actual order-book depth; they are not proof of any specific cause. No larger position sizing or real-money trading is authorized by this release.
