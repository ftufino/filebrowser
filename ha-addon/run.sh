#!/bin/sh
set -e

OPTIONS=/data/options.json

BASE_URL=$(grep -o '"base_url": *"[^"]*"' "$OPTIONS" | sed 's/.*: *"\(.*\)"/\1/' || true)

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

echo "Starting File Browser on port 8080 with base URL '${BASE_URL}'"
exec filebrowser --config /data/settings.json --port 8080 --baseURL "$BASE_URL"
