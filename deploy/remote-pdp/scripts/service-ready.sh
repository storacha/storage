#!/usr/bin/env bash
# Usage: service-ready.sh HOST PORT [TIMEOUT_SECONDS]
# Example: service-ready.sh 127.0.0.1 12300 60

HOST="${1}"
PORT="${2}"
TIMEOUT="${3:-60}"

for i in $(seq 1 "$TIMEOUT"); do
  if nc -z "$HOST" "$PORT" >/dev/null 2>&1; then
    exit 0
  fi
  sleep 1
done

echo "Service at $HOST:$PORT didn't become ready within $TIMEOUT seconds."
exit 1
