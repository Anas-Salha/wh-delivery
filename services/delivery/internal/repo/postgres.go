package repo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type Postgres struct {
	db *gorm.DB
}

func NewPostgres(db *gorm.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) CreateWebhook(ctx context.Context, webhook *Webhook) error {
	log.Printf("[repo] CreateWebhook client_id=%d", webhook.ClientID)
	return p.db.WithContext(ctx).Create(webhook).Error
}

func (p *Postgres) GetWebhook(ctx context.Context, id int64) (Webhook, error) {
	log.Printf("[repo] GetWebhook id=%d", id)
	var webhook Webhook
	if err := p.db.WithContext(ctx).First(&webhook, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Webhook{}, ErrNotFound
		}
		return Webhook{}, err
	}
	return webhook, nil
}

func (p *Postgres) ListWebhooksByClient(ctx context.Context, clientID int64) ([]Webhook, error) {
	log.Printf("[repo] ListWebhooksByClient client_id=%d", clientID)
	var results []Webhook
	if err := p.db.WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("created_at DESC").
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (p *Postgres) UpdateWebhook(ctx context.Context, webhook Webhook) error {
	log.Printf("[repo] UpdateWebhook id=%d", webhook.ID)
	updates := map[string]any{
		"callback_url": webhook.CallbackURL,
		"event_types":  webhook.EventTypes,
		"status":       webhook.Status,
		"retry_config": webhook.RetryConfig,
		"updated_at":   time.Now().UTC(),
	}

	result := p.db.WithContext(ctx).
		Model(&Webhook{}).
		Where("id = ?", webhook.ID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) DeleteWebhook(ctx context.Context, id int64) error {
	log.Printf("[repo] DeleteWebhook id=%d", id)
	result := p.db.WithContext(ctx).Delete(&Webhook{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) CreateEvent(ctx context.Context, event *Event) error {
	log.Printf("[repo] CreateEvent source_id=%d event_type=%s", event.SourceID, event.EventType)
	if err := p.db.WithContext(ctx).Create(event).Error; err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	return nil
}

func (p *Postgres) GetEvent(ctx context.Context, id int64) (Event, error) {
	log.Printf("[repo] GetEvent id=%d", id)
	var event Event
	if err := p.db.WithContext(ctx).First(&event, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Event{}, ErrNotFound
		}
		return Event{}, err
	}
	return event, nil
}

func (p *Postgres) GetEventByIdempotencyKey(ctx context.Context, sourceID int64, key string) (Event, error) {
	log.Printf("[repo] GetEventByIdempotencyKey source_id=%d", sourceID)
	var event Event
	if err := p.db.WithContext(ctx).
		First(&event, "source_id = ? AND idempotency_key = ?", sourceID, key).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Event{}, ErrNotFound
		}
		return Event{}, err
	}
	return event, nil
}

func (p *Postgres) CreateDelivery(ctx context.Context, delivery *Delivery) error {
	log.Printf("[repo] CreateDelivery event_id=%d webhook_id=%d status=%s", delivery.EventID, delivery.WebhookID, delivery.Status)
	return p.db.WithContext(ctx).Create(delivery).Error
}

func (p *Postgres) UpdateDeliveryAttempt(ctx context.Context, deliveryID int64, status string, attempts int, lastAttemptAt time.Time, lastError *string) error {
	log.Printf("[repo] UpdateDeliveryAttempt id=%d status=%s attempts=%d", deliveryID, status, attempts)
	updates := map[string]any{
		"status":          status,
		"attempts":        attempts,
		"last_attempt_at": lastAttemptAt,
		"last_error":      lastError,
	}

	result := p.db.WithContext(ctx).
		Model(&Delivery{}).
		Where("id = ?", deliveryID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) CreateSource(ctx context.Context, source *Source) error {
	log.Printf("[repo] CreateSource source_name=%s", source.SourceName)
	return p.db.WithContext(ctx).Create(source).Error
}

func (p *Postgres) GetSource(ctx context.Context, id int64) (Source, error) {
	log.Printf("[repo] GetSource id=%d", id)
	var source Source
	if err := p.db.WithContext(ctx).First(&source, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Source{}, ErrNotFound
		}
		return Source{}, err
	}
	return source, nil
}

func (p *Postgres) GetSourceByAPIKey(ctx context.Context, apiKey string) (Source, error) {
	log.Printf("[repo] GetSourceByAPIKey api_key=%s", apiKey)
	var source Source
	if err := p.db.WithContext(ctx).First(&source, "api_key = ?", apiKey).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Source{}, ErrNotFound
		}
		return Source{}, err
	}
	return source, nil
}

func (p *Postgres) UpdateSource(ctx context.Context, source Source) error {
	log.Printf("[repo] UpdateSource id=%d", source.ID)
	updates := map[string]any{
		"status":              source.Status,
		"allowed_event_types": source.AllowedEventTypes,
		"updated_at":          time.Now().UTC(),
	}

	result := p.db.WithContext(ctx).
		Model(&Source{}).
		Where("id = ?", source.ID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) DeleteSource(ctx context.Context, id int64) error {
	log.Printf("[repo] DeleteSource id=%d", id)
	result := p.db.WithContext(ctx).Delete(&Source{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return string(pqErr.Code) == "23505"
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
