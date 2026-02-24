package httpapi

type CreateWebhookRequest struct {
	CallbackURL       string         `json:"callback_url"`
	EventTypes        []string       `json:"event_types"`
	RetryConfig       map[string]any `json:"retry_config"`
	RateLimit         map[string]any `json:"rate_limit"`
}

type UpdateWebhookRequest struct {
	CallbackURL string         `json:"callback_url"`
	EventTypes  []string       `json:"event_types"`
	Status      string         `json:"status"`
	RetryConfig map[string]any `json:"retry_config"`
}

type CreateWebhookResponse struct {
	WebhookID           int64  `json:"webhook_id"`
	Status              string `json:"status"`
	SigningSecret       string `json:"signing_secret"`
	CreatedAt           string `json:"created_at"`
	CallbackURLVerified bool   `json:"callback_url_verified"`
}

type UpdateWebhookResponse struct {
	WebhookID int64  `json:"webhook_id"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}
