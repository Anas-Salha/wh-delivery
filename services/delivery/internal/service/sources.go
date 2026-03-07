package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"gorm.io/datatypes"

	"github.com/anas-salha/wh-delivery/delivery/internal/repo"
)

type SourceRepo interface {
	CreateSource(ctx context.Context, source *repo.Source) error
	GetSource(ctx context.Context, id int64) (repo.Source, error)
	GetSourceByAPIKey(ctx context.Context, apiKey string) (repo.Source, error)
	UpdateSource(ctx context.Context, source repo.Source) error
	DeleteSource(ctx context.Context, id int64) error
}

type EventRepo interface {
	CreateEvent(ctx context.Context, event *repo.Event) error
	GetEventByIdempotencyKey(ctx context.Context, sourceID int64, key string) (repo.Event, error)
}

type SourcesService struct {
	repo      SourceRepo
	eventRepo EventRepo
}

func NewSourcesService(repo SourceRepo, eventRepo EventRepo) *SourcesService {
	return &SourcesService{repo: repo, eventRepo: eventRepo}
}

func (s *SourcesService) Create(ctx context.Context, input CreateSourceInput) (repo.Source, error) {
	log.Printf(
		"[service] SourcesService.Create source_name=%q api_key=%q webhook_secret=%q allowed_event_types=%v status=%q",
		input.SourceName,
		redactValue(input.APIKey),
		redactValue(input.WebhookSecret),
		input.AllowedEventTypes,
		input.Status,
	)
	if input.Status == "" {
		input.Status = "active"
	}

	now := time.Now().UTC()
	source := repo.Source{
		SourceName:        input.SourceName,
		APIKey:            input.APIKey,
		WebhookSecret:     input.WebhookSecret,
		AllowedEventTypes: input.AllowedEventTypes,
		Status:            input.Status,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.CreateSource(ctx, &source); err != nil {
		return repo.Source{}, err
	}

	return source, nil
}

func (s *SourcesService) Update(ctx context.Context, id int64, input UpdateSourceInput) (repo.Source, error) {
	log.Printf(
		"[service] SourcesService.Update source_id=%d status=%q allowed_event_types=%v",
		id,
		input.Status,
		input.AllowedEventTypes,
	)
	current, err := s.repo.GetSource(ctx, id)
	if err != nil {
		return repo.Source{}, err
	}

	if input.Status != "" {
		current.Status = input.Status
	}
	if len(input.AllowedEventTypes) > 0 {
		current.AllowedEventTypes = input.AllowedEventTypes
	}

	current.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateSource(ctx, current); err != nil {
		return repo.Source{}, err
	}

	return current, nil
}

func (s *SourcesService) Delete(ctx context.Context, id int64) error {
	log.Printf("[service] SourcesService.Delete source_id=%d", id)
	return s.repo.DeleteSource(ctx, id)
}

func (s *SourcesService) GetByAPIKey(ctx context.Context, apiKey string) (repo.Source, error) {
	log.Printf("[service] SourcesService.GetByAPIKey api_key=%q", redactValue(apiKey))
	return s.repo.GetSourceByAPIKey(ctx, apiKey)
}

func (s *SourcesService) PushEvent(ctx context.Context, sourceID int64, input PushEventInput) (repo.Event, error) {
	log.Printf(
		"[service] SourcesService.PushEvent source_id=%d idempotency_key=%q event_type=%q occurred_at=%q",
		sourceID,
		input.IdempotencyKey,
		input.EventType,
		input.OccurredAt,
	)
	payload := map[string]any{
		"occurred_at": input.OccurredAt,
		"data":        input.Data,
		"metadata":    input.Metadata,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return repo.Event{}, err
	}

	event := repo.Event{
		SourceID:       sourceID,
		IdempotencyKey: input.IdempotencyKey,
		EventType:      input.EventType,
		Payload:        datatypes.JSON(raw),
		CreatedAt:      time.Now().UTC(),
	}

	if err := s.eventRepo.CreateEvent(ctx, &event); err != nil {
		if errors.Is(err, repo.ErrConflict) {
			existing, getErr := s.eventRepo.GetEventByIdempotencyKey(ctx, sourceID, input.IdempotencyKey)
			if getErr != nil {
				return repo.Event{}, getErr
			}
			return existing, nil
		}
		return repo.Event{}, err
	}

	return event, nil
}
