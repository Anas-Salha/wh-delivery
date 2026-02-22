package repo

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type Webhook struct {
	ID            int64          `gorm:"primaryKey;column:id"`
	ClientID      int64          `gorm:"column:client_id"`
	CallbackURL   string         `gorm:"column:callback_url"`
	SigningSecret string         `gorm:"column:signing_secret"`
	EventTypes    pq.StringArray `gorm:"type:text[];column:event_types"`
	Status        string         `gorm:"column:status"`
	RetryConfig   datatypes.JSON `gorm:"type:jsonb;column:retry_config"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
}

type Event struct {
	ID             int64          `gorm:"primaryKey;column:id"`
	SourceID       int64          `gorm:"column:source_id"`
	IdempotencyKey string         `gorm:"column:idempotency_key"`
	EventType      string         `gorm:"column:event_type"`
	Payload        datatypes.JSON `gorm:"type:jsonb;column:payload"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
}

type Delivery struct {
	ID            int64      `gorm:"primaryKey;column:id"`
	EventID       int64      `gorm:"column:event_id"`
	WebhookID     int64      `gorm:"column:webhook_id"`
	Status        string     `gorm:"column:status"`
	Attempts      int        `gorm:"column:attempts"`
	LastAttemptAt *time.Time `gorm:"column:last_attempt_at"`
	LastError     *string    `gorm:"column:last_error"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
}

type Source struct {
	ID                int64          `gorm:"primaryKey;column:id"`
	SourceName        string         `gorm:"column:source_name"`
	APIKey            string         `gorm:"column:api_key"`
	WebhookSecret     string         `gorm:"column:webhook_secret"`
	AllowedEventTypes pq.StringArray `gorm:"type:text[];column:allowed_event_types"`
	Status            string         `gorm:"column:status"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
}
