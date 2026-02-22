package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/anas-salha/wh-delivery/delivery/internal/repo"
	"github.com/anas-salha/wh-delivery/delivery/internal/service"
)

type SourcesService interface {
	Create(ctx context.Context, input service.CreateSourceInput) (repo.Source, error)
	Update(ctx context.Context, id int64, input service.UpdateSourceInput) (repo.Source, error)
	Delete(ctx context.Context, id int64) error
	PushEvent(ctx context.Context, sourceID int64, input service.PushEventInput) (repo.Event, error)
}

type SourcesHandler struct {
	serviceName string
	svc         SourcesService
}

func NewSourcesHandler(serviceName string, svc SourcesService) *SourcesHandler {
	return &SourcesHandler{serviceName: serviceName, svc: svc}
}

func (h *SourcesHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/sources", h.createSource)
	rg.PATCH("/sources/:source_id", h.updateSource)
	rg.DELETE("/sources/:source_id", h.deleteSource)
	rg.POST("/sources/:source_id/events", h.pushEvents)
}

func (h *SourcesHandler) createSource(c *gin.Context) {
	var req CreateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	input := service.CreateSourceInput{
		SourceName:        req.SourceName,
		APIKey:            req.APIKey,
		WebhookSecret:     req.WebhookSecret,
		AllowedEventTypes: req.AllowedEventTypes,
		Status:            "active",
	}

	source, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create source"})
		return
	}

	resp := CreateSourceResponse{
		SourceID:          source.ID,
		Status:            source.Status,
		SigningAlgo:       "hmac-sha256",
		AllowedEventTypes: source.AllowedEventTypes,
		CreatedAt:         source.CreatedAt.UTC().Format(time.RFC3339),
	}

	fmt.Printf("[%s] POST /api/v1/sources source_id=%d status=%s\n", h.serviceName, source.ID, source.Status)

	writeJSON(c, http.StatusCreated, resp)
}

func (h *SourcesHandler) updateSource(c *gin.Context) {
	var req UpdateSourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	sourceID, err := parseInt64Param(c.Param("source_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source_id"})
		return
	}

	input := service.UpdateSourceInput{
		Status:            req.Status,
		AllowedEventTypes: req.AllowedEventTypes,
	}

	source, err := h.svc.Update(c.Request.Context(), sourceID, input)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update source"})
		return
	}

	fmt.Printf("[%s] PATCH /api/v1/sources/%d auth=%q payload=%+v\n", h.serviceName, sourceID, c.GetHeader("Authorization"), req)

	resp := UpdateSourceResponse{
		SourceID:          source.ID,
		Status:            source.Status,
		AllowedEventTypes: source.AllowedEventTypes,
		UpdatedAt:         source.UpdatedAt.UTC().Format(time.RFC3339),
	}
	writeJSON(c, http.StatusOK, resp)
}

func (h *SourcesHandler) deleteSource(c *gin.Context) {
	sourceID, err := parseInt64Param(c.Param("source_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source_id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), sourceID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "source not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete source"})
		return
	}

	fmt.Printf("[%s] DELETE /api/v1/sources/%d auth=%q\n", h.serviceName, sourceID, c.GetHeader("Authorization"))
	c.Status(http.StatusNoContent)
}

func (h *SourcesHandler) pushEvents(c *gin.Context) {
	var req PushEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}
	if req.IdempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idempotency_key is required"})
		return
	}

	sourceID, err := parseInt64Param(c.Param("source_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source_id"})
		return
	}

	input := service.PushEventInput{
		IdempotencyKey: req.IdempotencyKey,
		EventType:      req.EventType,
		OccurredAt:     req.OccurredAt,
		Data:           req.Data,
		Metadata:       req.Metadata,
	}

	event, err := h.svc.PushEvent(c.Request.Context(), sourceID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to push event"})
		return
	}

	fmt.Printf(
		"[%s] POST /api/v1/sources/%d/events auth=%q signature=%q timestamp=%q payload=%+v\n",
		h.serviceName,
		sourceID,
		c.GetHeader("Authorization"),
		c.GetHeader("X-Source-Signature"),
		c.GetHeader("X-Source-Timestamp"),
		req,
	)

	resp := PushEventResponse{
		Accepted:   true,
		EventID:    event.ID,
		ReceivedAt: event.CreatedAt.UTC().Format(time.RFC3339),
	}
	writeJSON(c, http.StatusAccepted, resp)
}
