# Consumer API

POST /api/v1/webhooks
header: Authorization: Bearer {JWT_TOKEN}
```json
{
	"callback_url": "string (required, must be HTTPS)",
	"event_types": ["string"] (optional, e.g., ["payment_success", "order_updated"])
	"retry_config": {
		"max_retries": 5,
		"initial_delay_ms": 1000,
		"max_delay_ms": 30000,
		"backoff_multiplies": 2.0
	}
	"rate_limit": {
		"requests_per_minute": 100
	}
}
```

response:
200 (OK), 400 (Bad Request), 401 (Unauthorized), 5xx (server error)
```json
{
	"webhook_id": 123,
	"status": "active|inactive",
	"signing_secret": "string (server-generated for HMAC)",
	"created_at": "", //timestamp
	"callback_url_verified": "boolean"
}
```

PATCH /api/v1/webhooks/{webhook_id}
header: Authorization: Bearer {JWT_TOKEN}
```json
{
	"callback_url": "string (optional)",
	"event_types": ["string"] (optional),
	"status": "active|inactive" (optional),
	"retry_config": {...} (optional)
}
```

response:
200 (OK), 400 (Bad Request), 404 (Not Found), 401 (Unauthorized)
```json
{
	"webhook_id": 123,
	"status": "active|inactive",
	"updated_at": "" //ISO8601 timestamp
}
```

DELETE /api/v1/webhooks/{webhook_id}
Authorization: Bearer {JWT_TOKEN}
Response: 204 No Content
