package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/anas-salha/wh-delivery/delivery/internal/repo"
	"github.com/anas-salha/wh-delivery/delivery/internal/service"
)

type WebhooksService interface {
	Create(ctx context.Context, input service.CreateWebhookInput) (repo.Webhook, error)
	Update(ctx context.Context, id int64, input service.UpdateWebhookInput) (repo.Webhook, error)
	Delete(ctx context.Context, id int64) error
}

type WebhooksHandler struct {
	serviceName string
	svc         WebhooksService
}

func NewWebhooksHandler(serviceName string, svc WebhooksService) *WebhooksHandler {
	return &WebhooksHandler{serviceName: serviceName, svc: svc}
}

func (h *WebhooksHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/webhooks", h.createWebhook)
	rg.PATCH("/webhooks/:webhook_id", h.updateWebhook)
	rg.DELETE("/webhooks/:webhook_id", h.deleteWebhook)
}

func (h *WebhooksHandler) createWebhook(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	input := service.CreateWebhookInput{
		CallbackURL:   req.CallbackURL,
		EventTypes:    req.EventTypes,
		SigningSecret: "",
		Status:        "active",
		RetryConfig:   req.RetryConfig,
	}

	webhook, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create webhook"})
		return
	}

	fmt.Printf("[%s] POST /api/v1/webhooks auth=%q payload=%+v\n", h.serviceName, c.GetHeader("Authorization"), req)

	resp := CreateWebhookResponse{
		WebhookID:           webhook.ID,
		Status:              webhook.Status,
		SigningSecret:       webhook.SigningSecret,
		CreatedAt:           webhook.CreatedAt.UTC().Format(time.RFC3339),
		CallbackURLVerified: false,
	}

	writeJSON(c, http.StatusOK, resp)
}

func (h *WebhooksHandler) updateWebhook(c *gin.Context) {
	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	webhookID, err := parseInt64Param(c.Param("webhook_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook_id"})
		return
	}

	input := service.UpdateWebhookInput{
		CallbackURL: req.CallbackURL,
		EventTypes:  req.EventTypes,
		Status:      req.Status,
		RetryConfig: req.RetryConfig,
	}

	webhook, err := h.svc.Update(c.Request.Context(), webhookID, input)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update webhook"})
		return
	}

	fmt.Printf("[%s] PATCH /api/v1/webhooks/%d auth=%q payload=%+v\n", h.serviceName, webhookID, c.GetHeader("Authorization"), req)

	resp := UpdateWebhookResponse{
		WebhookID: webhook.ID,
		Status:    webhook.Status,
		UpdatedAt: webhook.UpdatedAt.UTC().Format(time.RFC3339),
	}
	writeJSON(c, http.StatusOK, resp)
}

func (h *WebhooksHandler) deleteWebhook(c *gin.Context) {
	webhookID, err := parseInt64Param(c.Param("webhook_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook_id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), webhookID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "webhook not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete webhook"})
		return
	}

	fmt.Printf("[%s] DELETE /api/v1/webhooks/%d auth=%q\n", h.serviceName, webhookID, c.GetHeader("Authorization"))
	c.Status(http.StatusNoContent)
}

func parseInt64Param(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}
