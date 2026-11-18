# SkewLab on Railway — k-signal-paper-lab

This deployment is **paper trading only**. It cannot place Polymarket live orders.

## Deployment

Create a separate `skewlab-paper` service connected to `twappleon/SkewLab` (main) in the existing `k-signal-paper-lab` project.

Attach a **persistent volume** mounted at `/app/data`. The Docker entrypoint starts with elevated privileges only to change that mount's directory ownership to the `app` user; both the Telegram smoke test and the long-running Go service run as `app`.

### Required Railway variables

```dotenv
DASHBOARD_USER=admin
DASHBOARD_PASSWORD=${{k-signal-paper-monitor.APP_PASSWORD}}
TELEGRAM_BOT_TOKEN=${{k-signal-binance-demo.TELEGRAM_BOT_TOKEN}}
TELEGRAM_CHAT_ID=${{k-signal-binance-demo.DEMO_TELEGRAM_CHAT_ID}}
SKEWLAB_TELEGRAM_TEST_ONCE=1
```

Never copy actual secrets or include them in GitHub. Use only permitted existing Railway reference variables. If `APP_PASSWORD` is not a strong secret, replace it in the Railway UI, not in chat.

Only generate a public Railway domain **after** setting dashboard credentials and confirming the service is healthy. The Go program uses HTTP Basic Auth for all paths except `/health`; Railway edge terminates HTTPS. Since the domain is accessible from the internet, protect credentials and avoid public sharing.

The startup test explicitly sends a single `🧪` Telegram message **only once**, after `/app/data` is mounted and the bot recipient is configured. After Telegram returns API HTTP 200 + `ok`, the marker `/app/data/.telegram-test-accepted` prevents repeated testing across container restarts. Telegram API acknowledgment does not prove a mobile notification was displayed; verify in the target chat. If the send fails, startup exits and Railway restarts, notifies an operator, and must not be described as success.

### Runtime verification

- Railway deployment status SUCCESS and `/health` HTTP 200: process is up (not market readiness).
- Passwordless Dashboard `/` must return 401 Unauthorized.
- Authenticated Dashboard `/` should return 200 through HTTPS.
- Inspect runtime logs for `Telegram test accepted by API` and `Telegram startup smoke test acknowledged and journaled`. Never include token-bearing URLs in logs.
- Check actual `/api/feeds` stream and storage stats for market readiness.
- Run one replica: a second instance would double-process paper observations.
- Record real 7–14 day paper results before considering live trading.
