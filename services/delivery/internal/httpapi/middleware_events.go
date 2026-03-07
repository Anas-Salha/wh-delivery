package httpapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/anas-salha/wh-delivery/delivery/internal/repo"
)

func (h *SourcesHandler) eventAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[httpapi] eventAuthMiddleware path=%s", c.Request.URL.Path)
		apiKey, ok := parseBearerToken(c.GetHeader("Authorization"))
		if !ok {
			log.Printf("[httpapi] eventAuthMiddleware missing bearer token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization"})
			c.Abort()
			return
		}

		signatureHeader := strings.TrimSpace(c.GetHeader("X-Source-Signature"))
		if signatureHeader == "" {
			log.Printf("[httpapi] eventAuthMiddleware missing signature")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing source signature"})
			c.Abort()
			return
		}

		timestampHeader := strings.TrimSpace(c.GetHeader("X-Source-Timestamp"))
		if timestampHeader == "" {
			log.Printf("[httpapi] eventAuthMiddleware missing timestamp")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing source timestamp"})
			c.Abort()
			return
		}
		if _, err := strconv.ParseInt(timestampHeader, 10, 64); err != nil {
			log.Printf("[httpapi] eventAuthMiddleware invalid timestamp=%q", timestampHeader)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source timestamp"})
			c.Abort()
			return
		}

		sourceID, err := parseInt64Param(c.Param("source_id"))
		if err != nil {
			log.Printf("[httpapi] eventAuthMiddleware invalid source_id=%q", c.Param("source_id"))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid source_id"})
			c.Abort()
			return
		}

		source, err := h.svc.GetByAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				log.Printf("[httpapi] eventAuthMiddleware unknown api_key=%q", redactValue(apiKey))
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
				c.Abort()
				return
			}
			log.Printf("[httpapi] eventAuthMiddleware auth failed api_key=%q err=%v", redactValue(apiKey), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "auth failed"})
			c.Abort()
			return
		}

		if source.ID != sourceID {
			log.Printf("[httpapi] eventAuthMiddleware api key mismatch source_id=%d expected=%d", sourceID, source.ID)
			c.JSON(http.StatusForbidden, gin.H{"error": "api key does not match source"})
			c.Abort()
			return
		}
		if strings.ToLower(source.Status) != "active" {
			log.Printf("[httpapi] eventAuthMiddleware inactive source_id=%d status=%q", source.ID, source.Status)
			c.JSON(http.StatusForbidden, gin.H{"error": "source is inactive"})
			c.Abort()
			return
		}

		body, err := c.GetRawData()
		if err != nil {
			log.Printf("[httpapi] eventAuthMiddleware read body error=%v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		if !validSourceSignature(signatureHeader, source.WebhookSecret, body) {
			log.Printf("[httpapi] eventAuthMiddleware invalid signature source_id=%d", source.ID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			c.Abort()
			return
		}

		log.Printf(
			"[httpapi] eventAuthMiddleware authorized source_id=%d api_key=%q timestamp=%q",
			source.ID,
			redactValue(apiKey),
			timestampHeader,
		)
		c.Set("source", source)
		c.Next()
	}
}

func validSourceSignature(header, secret string, body []byte) bool {
	expected := computeHMACSHA256(secret, body)
	signature := strings.TrimSpace(header)
	if signature == "" {
		return false
	}
	return hmac.Equal([]byte(signature), []byte(expected))
}

func computeHMACSHA256(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
