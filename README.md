# Webhook Delivery POC

Three Go microservices implementing the ingestion side of a webhook delivery system.

## Services

- **auth** (`:8081`): Issues RS256 JWTs for admin and user roles
- **delivery** (`:8080`): Manages sources and webhooks; ingests and stores events
- **consumer**: Heartbeat stub — actual delivery logic not yet implemented

## What's implemented

- Source management (admin-only): create/update/delete sources with API keys and webhook secrets
- Event ingestion: authenticated via API key + HMAC-SHA256 signature, with idempotency
- Webhook CRUD: register/update/delete webhook endpoints with event type filters
- Admin auth middleware: RS256 JWT validation (issuer `auth-service`, audience `delivery-service`, role `admin`)

## What's not yet implemented

- Delivery worker: no logic dispatches events to registered webhooks
- Webhook auth: webhook endpoints have no authentication
- Callback URL verification: always returns `false`
- Retry logic and rate limiting: fields stored but unused

## Run

```bash
docker compose up --build
```

Migrations run automatically on startup. Stop with `Ctrl+C`, clean up with:

```bash
docker compose down
```

## Source API

Get an admin token:

```bash
TOKEN=$(curl -s -X POST http://localhost:8081/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')
```

Create a source:

```bash
curl -i -X POST http://localhost:8080/api/v1/sources \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{
    "source_name": "payments",
    "api_key": "key-abc123",
    "webhook_secret": "secret-xyz",
    "allowed_event_types": ["payment.success", "payment.failed"]
  }'
```

Push an event (HMAC-SHA256 of the raw request body using the source's `webhook_secret`):

```bash
BODY='{"idempotency_key":"evt-1","event_type":"payment.success","data":{"amount":99}}'
SIG=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "secret-xyz" | awk '{print $2}')

curl -i -X POST http://localhost:8080/api/v1/sources/1/events \
  -H 'Authorization: Bearer key-abc123' \
  -H "X-Source-Signature: $SIG" \
  -H "X-Source-Timestamp: $(date +%s)" \
  -H 'Content-Type: application/json' \
  -d "$BODY"
```

## Webhook API

Create a webhook:

```bash
curl -i -X POST http://localhost:8080/api/v1/webhooks \
  -H 'Content-Type: application/json' \
  -d '{
    "callback_url": "https://example.com/webhook",
    "event_types": ["payment.success"]
  }'
```

Update a webhook:

```bash
curl -i -X PATCH http://localhost:8080/api/v1/webhooks/1 \
  -H 'Content-Type: application/json' \
  -d '{"status": "inactive"}'
```

Delete a webhook:

```bash
curl -i -X DELETE http://localhost:8080/api/v1/webhooks/1
```
