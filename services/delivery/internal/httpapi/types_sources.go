package httpapi

type CreateSourceRequest struct {
	SourceName        string   `json:"source_name"`
	APIKey            string   `json:"api_key"`
	WebhookSecret     string   `json:"webhook_secret"`
	AllowedEventTypes []string `json:"allowed_event_types"`
}

type CreateSourceResponse struct {
	SourceID          int64    `json:"source_id"`
	Status            string   `json:"status"`
	SigningAlgo       string   `json:"signing_algo"`
	AllowedEventTypes []string `json:"allowed_event_types"`
	CreatedAt         string   `json:"created_at"`
}

type UpdateSourceRequest struct {
	Status            string   `json:"status"`
	AllowedEventTypes []string `json:"allowed_event_types"`
}

type UpdateSourceResponse struct {
	SourceID          int64    `json:"source_id"`
	Status            string   `json:"status"`
	AllowedEventTypes []string `json:"allowed_event_types"`
	UpdatedAt         string   `json:"updated_at"`
}

type PushEventRequest struct {
	IdempotencyKey string         `json:"idempotency_key"`
	EventType      string         `json:"event_type"`
	OccurredAt     string         `json:"occurred_at"`
	Data           map[string]any `json:"data"`
	Metadata       map[string]any `json:"metadata"`
}

type PushEventResponse struct {
	Accepted   bool   `json:"accepted"`
	EventID    int64  `json:"event_id"`
	ReceivedAt string `json:"received_at"`
}
