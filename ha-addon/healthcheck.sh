#!/bin/sh
PORT=$(grep -o '"port": *[0-9]*' /data/options.json | grep -o '[0-9]*$')
nc -z 127.0.0.1 "${PORT:-8080}"
