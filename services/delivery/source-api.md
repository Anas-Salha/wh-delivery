# Source API

## Source onboarding

POST /api/v1/sources
header: Authorization: Bearer {ADMIN_JWT}
```json
{
  "source_name": "string (required)",
  "api_key": "string (required, externally generated)",
  "webhook_secret": "string (required, HMAC secret, externally generated)",
  "allowed_event_types": ["string"]
}
```

response:
201 (Created), 400 (Bad Request), 401 (Unauthorized), 5xx (server error)
```json
{
  "source_id": 123,
  "status": "active|inactive",
  "signing_algo": "hmac-sha256",
  "allowed_event_types": ["string"],
  "created_at": "" //ISO8601 timestamp
}
```

## Update source

PATCH /api/v1/sources/{source_id}
header: Authorization: Bearer {ADMIN_JWT}
```json
{
  "status": "active|inactive (optional)",
  "allowed_event_types": ["string"] //optional
}
```

response:
200 (OK), 400 (Bad Request), 401 (Unauthorized), 404 (Not Found), 5xx (server error)
```json
{
  "source_id": 123,
  "status": "active|inactive",
  "allowed_event_types": ["string"],
  "updated_at": "" //ISO8601 timestamp
}
```

## Delete source

DELETE /api/v1/sources/{source_id}
header: Authorization: Bearer {ADMIN_JWT}
Response: 204 No Content

## Push events
POST /api/v1/sources/{source_id}/events
header: Authorization: Bearer {SOURCE_API_KEY}
header: X-Source-Signature:  <signature> //sha256=HMAC_SHA256(secret, request_body)
header: X-Source-Timestamp: 1705312200
```json
{
  "idempotency_key": "string (required)",
  "event_type": "string (required)",
  "occurred_at": "" , //ISO8601 timestamp
  "data": {
    "any": "payload"
  },
  "metadata": {
    "trace_id": "string",
    "source_version": "string"
  }
}
```

response:
202 (Accepted), 400 (Bad Request), 401 (Unauthorized), 404 (Not Found), 5xx (server error)
```json
{
  "accepted": true,
  "event_id": 456,
  "received_at": "" //ISO8601 timestamp
}
```
