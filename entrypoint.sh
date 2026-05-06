#!/bin/sh
set -e
chown eagle:eagle /data
exec su-exec eagle /app/eagle-bank
