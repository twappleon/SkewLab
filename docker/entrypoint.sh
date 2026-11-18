#!/bin/sh
# Railway mounts a fresh volume as root. Fix only the mount directory,
# then drop root privileges for Telegram verification and the paper bot.
set -eu

if [ ! -d /app/data ]; then
  echo "Required persistent volume /app/data missing" >&2
  exit 1
fi

chown app:app /app/data
chmod 0700 /app/data

if [ "${SKEWLAB_TELEGRAM_TEST_ONCE:-0}" = "1" ] && [ ! -f /app/data/.telegram-test-accepted ]; then
  if su-exec app /app/telegram-test --send; then
    su-exec app sh -c 'umask 077; : > /app/data/.telegram-test-accepted'
    echo "Telegram startup smoke test acknowledged and journaled"
  else
    echo "Telegram startup smoke test failed; refusing to claim ready" >&2
    exit 1
  fi
fi

exec su-exec app /app/skewbot --data-dir /app/data
