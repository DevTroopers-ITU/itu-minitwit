#!/bin/sh
if [ -f /run/secrets/discord_webhook_url ]; then
  export DISCORD_WEBHOOK_URL=$(cat /run/secrets/discord_webhook_url)
fi
exec /run.sh "$@"
