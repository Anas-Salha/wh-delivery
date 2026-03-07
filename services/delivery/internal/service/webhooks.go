package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gorm.io/datatypes"

	"github.com/anas-salha/wh-delivery/delivery/internal/repo"
)

type WebhookRepo interface {
	CreateWebhook(ctx context.Context, webhook *repo.Webhook) error
	GetWebhook(ctx context.Context, id int64) (repo.Webhook, error)
	UpdateWebhook(ctx context.Context, webhook repo.Webhook) error
	DeleteWebhook(ctx context.Context, id int64) error
}

type WebhooksService struct {
	repo WebhookRepo
}

func NewWebhooksService(repo WebhookRepo) *WebhooksService {
	return &WebhooksService{repo: repo}
}

func (s *WebhooksService) Create(ctx context.Context, input CreateWebhookInput) (repo.Webhook, error) {
	log.Printf(
		"[service] WebhooksService.Create client_id=%d callback_url=%q event_types=%v signing_secret=%q status=%q",
		input.ClientID,
		input.CallbackURL,
		input.EventTypes,
		redactValue(input.SigningSecret),
		input.Status,
	)
	if input.ClientID == 0 {
		input.ClientID = 1
	}
	if input.SigningSecret == "" {
		input.SigningSecret = "whsec_dev"
	}
	if input.Status == "" {
		input.Status = "active"
	}

	var retryJSON datatypes.JSON
	if input.RetryConfig != nil {
		raw, err := json.Marshal(input.RetryConfig)
		if err != nil {
			return repo.Webhook{}, err
		}
		retryJSON = datatypes.JSON(raw)
	}

	now := time.Now().UTC()
	webhook := repo.Webhook{
		ClientID:      input.ClientID,
		CallbackURL:   input.CallbackURL,
		SigningSecret: input.SigningSecret,
		EventTypes:    input.EventTypes,
		Status:        input.Status,
		RetryConfig:   retryJSON,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateWebhook(ctx, &webhook); err != nil {
		return repo.Webhook{}, err
	}

	return webhook, nil
}

func (s *WebhooksService) Update(ctx context.Context, id int64, input UpdateWebhookInput) (repo.Webhook, error) {
	log.Printf(
		"[service] WebhooksService.Update webhook_id=%d callback_url=%q event_types=%v status=%q",
		id,
		input.CallbackURL,
		input.EventTypes,
		input.Status,
	)
	current, err := s.repo.GetWebhook(ctx, id)
	if err != nil {
		return repo.Webhook{}, err
	}

	if input.CallbackURL != "" {
		current.CallbackURL = input.CallbackURL
	}
	if len(input.EventTypes) > 0 {
		current.EventTypes = input.EventTypes
	}
	if input.Status != "" {
		current.Status = input.Status
	}
	if input.RetryConfig != nil {
		raw, err := json.Marshal(input.RetryConfig)
		if err != nil {
			return repo.Webhook{}, err
		}
		current.RetryConfig = datatypes.JSON(raw)
	}

	current.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateWebhook(ctx, current); err != nil {
		return repo.Webhook{}, err
	}

	return current, nil
}

func (s *WebhooksService) Delete(ctx context.Context, id int64) error {
	log.Printf("[service] WebhooksService.Delete webhook_id=%d", id)
	return s.repo.DeleteWebhook(ctx, id)
}
