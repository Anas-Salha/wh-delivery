#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required for pretty printing. Install jq and retry." >&2
  exit 1
fi

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

LAST_BODY=""

print_step() {
  printf "\n==> %s\n" "$1"
}

request() {
  local label="$1"
  local method="$2"
  local url="$3"
  local body="$4"
  shift 4 || true

  local headers_file="$tmp_dir/headers"
  local body_file="$tmp_dir/body"

  : >"$headers_file"
  : >"$body_file"

  print_step "$label"

  local -a cmd=(curl -sS -D "$headers_file" -o "$body_file" -X "$method" "$url")
  for header in "$@"; do
    cmd+=(-H "$header")
  done
  if [[ -n "$body" ]]; then
    cmd+=(-d "$body")
  fi

  "${cmd[@]}"
  cat "$headers_file"

  if [[ -s "$body_file" ]]; then
    jq . "$body_file"
    LAST_BODY="$(cat "$body_file")"
  else
    printf "\n"
    LAST_BODY=""
  fi
}

request "Create source" POST "$BASE_URL/api/v1/sources" \
  '{
    "source_name": "stripe_payment_processor",
    "api_key": "sk_live_stripe_provided_key",
    "webhook_secret": "whsec_stripe_provided_secret",
    "allowed_event_types": ["payment_intent.succeeded", "payment_intent.failed"]
  }' \
  "Authorization: Bearer admin-token" \
  "Content-Type: application/json"

source_id=$(echo "$LAST_BODY" | jq -r '.source_id')

request "Update source" PATCH "$BASE_URL/api/v1/sources/$source_id" \
  '{
    "status": "active",
    "allowed_event_types": ["payment_intent.succeeded"]
  }' \
  "Authorization: Bearer admin-token" \
  "Content-Type: application/json"

BODY='{
  "idempotency_key": "evt_001",
  "event_type": "payment_intent.succeeded",
  "occurred_at": "2026-02-19T23:00:00Z",
  "data": {"amount": 1250, "currency": "USD"},
  "metadata": {"trace_id": "trace_abc"}
}'
SIGNATURE="sha256=dummy"
TIMESTAMP="1705312200"

request "Push event" POST "$BASE_URL/api/v1/sources/$source_id/events" \
  "$BODY" \
  "Authorization: Bearer source-token" \
  "X-Source-Signature: $SIGNATURE" \
  "X-Source-Timestamp: $TIMESTAMP" \
  "Content-Type: application/json"

request "Create webhook" POST "$BASE_URL/api/v1/webhooks" \
  '{
    "callback_url": "https://example.com/webhook",
    "event_types": ["payment_intent.succeeded"],
    "trigger_conditions": {"filters": {"amount_gt": 1000}},
    "retry_config": {"max_retries": 5},
    "rate_limit": {"requests_per_minute": 100}
  }' \
  "Authorization: Bearer consumer-token" \
  "Content-Type: application/json"

webhook_id=$(echo "$LAST_BODY" | jq -r '.webhook_id')

request "Update webhook" PATCH "$BASE_URL/api/v1/webhooks/$webhook_id" \
  '{
    "callback_url": "https://example.com/new",
    "status": "active"
  }' \
  "Authorization: Bearer consumer-token" \
  "Content-Type: application/json"

request "Delete webhook" DELETE "$BASE_URL/api/v1/webhooks/$webhook_id" \
  "" \
  "Authorization: Bearer consumer-token"

request "Delete source" DELETE "$BASE_URL/api/v1/sources/$source_id" \
  "" \
  "Authorization: Bearer admin-token"
