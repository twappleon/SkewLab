# SkewLab v0.5.0 — Performance methodology

- Calendar windows are Asia/Phnom_Penh (UTC+07:00), including today, over 7 and 14 local calendar dates.
- Only CLOSED/SETTLED positions with an exit time count toward realized performance. OPEN and PENDING_RESOLUTION inventory is excluded.
- Net P&L is the sum of modeled realized values; it is not real or withdrawable money.
- Profit Factor = gross positive closes / magnitude of gross negative closes. Undefined when there are zero observed losses, returned as JSON null rather than infinity.
- Average net per close is period realized P&L divided by recorded closed transactions. No close means null.
- Win rate counts strictly positive closed-position outcomes; a flat zero P&L closure counts as non-win.
- Incomplete or conflicting historical ledger close counts are flagged with data_complete:false and sample_level:incomplete_history.
- 0–29 closes/14d = insufficient; 30–99 = preliminary; >=100 = research sample only. These thresholds do NOT prove profitability.
- Public aggregate pages: /, /performance, /api/performance. Private endpoints remain HTTP Basic protected. Do not remove the Railway DASHBOARD_PASSWORD.
- Daily Telegram summary keeps the original daily date-based deduplication key and adds 7/14 day aggregate results.
- Full trading, oracle data and market execution are unchanged; real-money orders remain disabled.
