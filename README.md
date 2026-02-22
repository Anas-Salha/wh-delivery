# Webhook Delivery POC (Skeleton)

This repository contains three minimal Go services to validate an end-to-end webhook delivery flow.

## Services
- `source`: event producer placeholder
- `delivery`: webhook delivery system placeholder
- `consumer`: webhook receiver placeholder

## Run

```bash
docker compose up --build
```

Run migrations (in a separate terminal):

```bash
./scripts/migrate.sh
```

Stop with `Ctrl+C`, then clean up with:

```bash
docker compose down
```

## Consumer API quick test

Create a webhook:

```bash
curl -i -X POST http://localhost:8080/api/v1/webhooks \
  -H 'Authorization: Bearer test-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "callback_url": "https://example.com/webhook",
    "event_types": ["payment_success"],
    "trigger_conditions": {"filters": {"amount_gt": 1000}},
    "retry_config": {"max_retries": 5},
    "rate_limit": {"requests_per_minute": 100}
  }'
```

Update a webhook:

```bash
curl -i -X PATCH http://localhost:8080/api/v1/webhooks/wh_123 \
  -H 'Authorization: Bearer test-token' \
  -H 'Content-Type: application/json' \
  -d '{
    "callback_url": "https://example.com/new",
    "status": "active"
  }'
```

Delete a webhook:

```bash
curl -i -X DELETE http://localhost:8080/api/v1/webhooks/wh_123 \
  -H 'Authorization: Bearer test-token'
```
