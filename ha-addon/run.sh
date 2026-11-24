#!/bin/sh
set -e

if [ ! -f /data/settings.json ]; then
  cat > /data/settings.json <<'JSON'
{
  "port": 8080,
  "address": "0.0.0.0",
  "baseURL": "",
  "log": "stdout",
  "database": "/data/filebrowser.db",
  "root": "/share"
}
JSON
fi

exec filebrowser --config /data/settings.json
