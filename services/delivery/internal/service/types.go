package service

type CreateWebhookInput struct {
	ClientID      int64
	CallbackURL   string
	SigningSecret string
	EventTypes    []string
	Status        string
	RetryConfig   map[string]any
}

type UpdateWebhookInput struct {
	CallbackURL string
	EventTypes  []string
	Status      string
	RetryConfig map[string]any
}

type CreateSourceInput struct {
	SourceName        string
	APIKey            string
	WebhookSecret     string
	AllowedEventTypes []string
	Status            string
}

type UpdateSourceInput struct {
	Status            string
	AllowedEventTypes []string
}

type PushEventInput struct {
	IdempotencyKey string
	EventType      string
	OccurredAt     string
	Data           map[string]any
	Metadata       map[string]any
}
