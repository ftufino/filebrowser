#!/bin/sh
set -e

OPTIONS=/data/options.json

PORT=$(grep -o '"port": *[0-9]*' "$OPTIONS" | grep -o '[0-9]*$' || true)
BASE_URL=$(grep -o '"base_url": *"[^"]*"' "$OPTIONS" | sed 's/.*: *"\(.*\)"/\1/' || true)

PORT=${PORT:-8080}
# Normalize to "/prefix" (no trailing slash); empty means served at the root.
BASE_URL=$(echo "$BASE_URL" | sed 's#^/*#/#; s#/*$##')
[ "$BASE_URL" = "/" ] && BASE_URL=""

if [ ! -f /data/settings.json ]; then
  cat > /data/settings.json <<'JSON'
{
  "address": "0.0.0.0",
  "log": "stdout",
  "database": "/data/filebrowser.db",
  "root": "/share"
}
JSON
fi

echo "Starting File Browser on port ${PORT} with base URL '${BASE_URL}'"
exec filebrowser --config /data/settings.json --port "$PORT" --baseurl "$BASE_URL"
