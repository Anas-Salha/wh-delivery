package delivery_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type createSourceRequest struct {
	SourceName        string   `json:"source_name"`
	APIKey            string   `json:"api_key"`
	WebhookSecret     string   `json:"webhook_secret"`
	AllowedEventTypes []string `json:"allowed_event_types"`
}

type createSourceResponse struct {
	SourceID int64  `json:"source_id"`
	Status   string `json:"status"`
}

type pushEventRequest struct {
	IdempotencyKey string         `json:"idempotency_key"`
	EventType      string         `json:"event_type"`
	OccurredAt     string         `json:"occurred_at"`
	Data           map[string]any `json:"data"`
	Metadata       map[string]any `json:"metadata"`
}

type pushEventResponse struct {
	Accepted bool  `json:"accepted"`
	EventID  int64 `json:"event_id"`
}

func TestSources_OnboardAndPushEvent(t *testing.T) {
	baseURL := deliveryBaseURL()

	adminToken := fetchAdminJWT(t)
	suffix := uniqueSuffix()
	sourceReq := createSourceRequest{
		SourceName:        "source-alpha-" + suffix,
		APIKey:            "source-api-key-" + suffix,
		WebhookSecret:     "webhook-secret-" + suffix,
		AllowedEventTypes: []string{"order.created"},
	}

	createResp := postJSON(t, baseURL+"/api/v1/sources", sourceReq, map[string]string{
		"Authorization": "Bearer " + adminToken,
	}, http.StatusCreated)

	var created createSourceResponse
	require.NoError(t, json.Unmarshal(createResp, &created))
	require.NotZero(t, created.SourceID)
	require.Equal(t, "active", created.Status)

	eventReq := pushEventRequest{
		IdempotencyKey: "idem-1-" + suffix,
		EventType:      "order.created",
		OccurredAt:     "2024-01-01T00:00:00Z",
		Data:           map[string]any{"order_id": "o-100"},
		Metadata:       map[string]any{"trace_id": "trace-1"},
	}

	eventResp := postSignedEvent(
		t,
		baseURL+"/api/v1/sources/"+strconv.FormatInt(created.SourceID, 10)+"/events",
		eventReq,
		sourceReq.APIKey,
		sourceReq.WebhookSecret,
		http.StatusAccepted,
	)

	var pushed pushEventResponse
	require.NoError(t, json.Unmarshal(eventResp, &pushed))
	require.True(t, pushed.Accepted)
	require.NotZero(t, pushed.EventID)
}

func TestSources_PushEvent_Idempotency(t *testing.T) {
	baseURL := deliveryBaseURL()

	adminToken := fetchAdminJWT(t)
	suffix := uniqueSuffix()
	sourceReq := createSourceRequest{
		SourceName:        "source-beta-" + suffix,
		APIKey:            "beta-api-key-" + suffix,
		WebhookSecret:     "beta-secret-" + suffix,
		AllowedEventTypes: []string{"order.created"},
	}

	createResp := postJSON(t, baseURL+"/api/v1/sources", sourceReq, map[string]string{
		"Authorization": "Bearer " + adminToken,
	}, http.StatusCreated)

	var created createSourceResponse
	require.NoError(t, json.Unmarshal(createResp, &created))

	eventReq := pushEventRequest{
		IdempotencyKey: "idem-dup-" + suffix,
		EventType:      "order.created",
		OccurredAt:     "2024-01-02T00:00:00Z",
		Data:           map[string]any{"order_id": "o-200"},
		Metadata:       map[string]any{"trace_id": "trace-2"},
	}

	first := postSignedEvent(
		t,
		baseURL+"/api/v1/sources/"+strconv.FormatInt(created.SourceID, 10)+"/events",
		eventReq,
		sourceReq.APIKey,
		sourceReq.WebhookSecret,
		http.StatusAccepted,
	)
	var firstResp pushEventResponse
	require.NoError(t, json.Unmarshal(first, &firstResp))

	second := postSignedEvent(
		t,
		baseURL+"/api/v1/sources/"+strconv.FormatInt(created.SourceID, 10)+"/events",
		eventReq,
		sourceReq.APIKey,
		sourceReq.WebhookSecret,
		http.StatusAccepted,
	)
	var secondResp pushEventResponse
	require.NoError(t, json.Unmarshal(second, &secondResp))

	require.Equal(t, firstResp.EventID, secondResp.EventID)
}

func fetchAdminJWT(t *testing.T) string {
	t.Helper()

	baseURL := os.Getenv("AUTH_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}

	body, err := json.Marshal(map[string]string{
		"username": "admin",
		"password": "admin123",
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, baseURL+"/admin/login", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	require.NotEmpty(t, payload.AccessToken)
	return payload.AccessToken
}

func deliveryBaseURL() string {
	baseURL := os.Getenv("DELIVERY_BASE_URL")
	if baseURL == "" {
		return "http://localhost:8080"
	}
	return baseURL
}

func uniqueSuffix() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func postJSON(t *testing.T, url string, payload any, headers map[string]string, expectedStatus int) []byte {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, expectedStatus, resp.StatusCode, "response body: %s", string(bodyBytes))
	return bodyBytes
}

func postSignedEvent(
	t *testing.T,
	url string,
	payload pushEventRequest,
	apiKey string,
	secret string,
	expectedStatus int,
) []byte {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	signature := computeHMACSHA256(secret, body)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Source-Signature", signature)
	req.Header.Set("X-Source-Timestamp", timestamp)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, expectedStatus, resp.StatusCode, "response body: %s", string(bodyBytes))
	return bodyBytes
}

func computeHMACSHA256(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
