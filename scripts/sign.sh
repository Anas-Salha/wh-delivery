#!/usr/bin/env bash

# Usage:
# ./sign.sh <secret> <body_file>
# or
# echo '{"foo":"bar"}' | ./sign.sh <secret>

set -e

SECRET="$1"

if [ -z "$SECRET" ]; then
  echo "Usage: $0 <secret> [body_file]"
  exit 1
fi

if [ -n "$2" ]; then
  BODY_CONTENT=$(cat "$2")
else
  BODY_CONTENT=$(cat)
fi

SIGNATURE=$(printf '%s' "$BODY_CONTENT" | \
  openssl dgst -sha256 -hmac "$SECRET" -binary | \
  xxd -p -c 256)

echo "$SIGNATURE"
