package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/db"
	"github.com/whg517/node-pulse/pulse/internal/models"
	"github.com/whg517/node-pulse/pulse/internal/security"
)

// PushService handles webhook push operations
type PushService struct {
	webhookQuerier     db.WebhookQuerier
	webhookLogsQuerier db.WebhookLogsQuerier
	httpClient         *http.Client
	baseURL            string
	urlValidator       func(string) error
}

// NewPushService creates a new PushService
func NewPushService(
	webhookQuerier db.WebhookQuerier,
	webhookLogsQuerier db.WebhookLogsQuerier,
	baseURL string,
) *PushService {
	return &PushService{
		webhookQuerier:     webhookQuerier,
		webhookLogsQuerier: webhookLogsQuerier,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:      baseURL,
		urlValidator: security.ValidateWebhookURL,
	}
}

// noopURLValidator is a validator that always passes (used in tests).
var noopURLValidator = func(string) error { return nil }

// WithURLValidator returns a new PushService with a custom URL validator.
// Pass nil to disable URL validation (useful in tests with http:// servers).
func (s *PushService) WithURLValidator(fn func(string) error) *PushService {
	if fn == nil {
		s.urlValidator = noopURLValidator
	} else {
		s.urlValidator = fn
	}
	return s
}

// SendWebhook sends an alert to a single webhook with retry logic
func (s *PushService) SendWebhook(ctx context.Context, alertEvent *models.AlertEvent, webhook *models.Webhook) error {
	maxRetries := 3
	backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff before retry
			select {
			case <-time.After(backoffs[attempt-1]):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Send webhook
		err := s.sendHTTP(ctx, alertEvent, webhook)
		if err == nil {
			// Success - log successful delivery with actual retry count
			s.logWebhookDelivery(ctx, alertEvent, webhook, "success", attempt, "")
			return nil
		}

		lastErr = err

		if attempt < maxRetries {
			slog.Warn("Webhook delivery failed, retrying",
				"webhook_id", webhook.ID,
				"attempt", attempt+1,
				"max_retries", maxRetries+1,
				"error", err)
		}
	}

	// All retries exhausted - log final failure
	s.logWebhookDelivery(ctx, alertEvent, webhook, "failure", maxRetries, lastErr.Error())
	return fmt.Errorf("webhook delivery failed after %d attempts: %w", maxRetries+1, lastErr)
}

// SendAlert sends an alert to all enabled webhooks concurrently
func (s *PushService) SendAlert(ctx context.Context, alertEvent *models.AlertEvent) error {
	// Get all enabled webhooks
	webhooks, err := s.webhookQuerier.GetWebhooks(ctx)
	if err != nil {
		return fmt.Errorf("failed to get webhooks: %w", err)
	}

	// Filter enabled webhooks
	var enabledWebhooks []*models.Webhook
	for _, webhook := range webhooks {
		if webhook.Enabled {
			enabledWebhooks = append(enabledWebhooks, webhook)
		}
	}

	if len(enabledWebhooks) == 0 {
		slog.Debug("No enabled webhooks configured, skipping webhook push")
		return nil
	}

	slog.Info("Sending alert to webhooks",
		"alert_id", alertEvent.ID,
		"webhook_count", len(enabledWebhooks))

	// Send to all webhooks concurrently
	var wg sync.WaitGroup
	errChan := make(chan error, len(enabledWebhooks))

	for _, webhook := range enabledWebhooks {
		wg.Add(1)
		go func(wh *models.Webhook) {
			defer wg.Done()
			err := s.SendWebhook(ctx, alertEvent, wh)
			errChan <- err
		}(webhook)
	}

	wg.Wait()
	close(errChan)

	// Collect all errors
	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
			slog.Error("Webhook delivery failed",
				"alert_id", alertEvent.ID,
				"error", err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("webhook push completed with %d failures out of %d webhooks", len(errors), len(enabledWebhooks))
	}

	slog.Info("Webhook push completed",
		"alert_id", alertEvent.ID,
		"webhook_count", len(enabledWebhooks),
		"success_count", len(enabledWebhooks)-len(errors))

	return nil
}

// sendHTTP sends a single HTTP POST request to a webhook
func (s *PushService) sendHTTP(ctx context.Context, alertEvent *models.AlertEvent, webhook *models.Webhook) error {
	// SSRF Protection: Validate webhook URL before sending
	validate := s.urlValidator
	if err := validate(webhook.URL); err != nil {
		slog.Warn("Webhook URL failed SSRF validation",
			"webhook_id", webhook.ID,
			"url", webhook.URL,
			"error", err)
		return fmt.Errorf("webhook URL validation failed: %w", err)
	}

	// Parse and validate URL structure
	parsedURL, err := url.Parse(webhook.URL)
	if err != nil {
		return fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Additional security: Ensure URL has required components
	if parsedURL.Host == "" {
		return fmt.Errorf("webhook URL missing host")
	}

	// Format alert event according to webhook's event format or default
	payload, err := s.formatAlertEvent(alertEvent, webhook)
	if err != nil {
		return fmt.Errorf("failed to format alert event: %w", err)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request with timeout
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// formatAlertEvent formats an alert event according to webhook's event_format or default template
func (s *PushService) formatAlertEvent(alertEvent *models.AlertEvent, webhook *models.Webhook) (map[string]any, error) {
	eventFormat := webhook.EventFormat
	if len(eventFormat) == 0 {
		eventFormat = models.DefaultEventFormat
	}

	formatted, ok := renderTemplateValue(eventFormat, s.alertTemplateValues(alertEvent)).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("webhook event format must be a JSON object")
	}
	return formatted, nil
}

func (s *PushService) alertTemplateValues(alertEvent *models.AlertEvent) map[string]any {
	return map[string]any{
		"AlertID":      alertEvent.ID,
		"Metric":       alertEvent.Metric,
		"Threshold":    alertEvent.Threshold,
		"CurrentValue": alertEvent.CurrentValue,
		"Level":        alertEvent.Level,
		"NodeID":       alertEvent.NodeID,
		// Alert events do not currently persist node names; keep this variable useful
		// for existing templates by falling back to the stable node identifier.
		"NodeName":    alertEvent.NodeID,
		"TriggeredAt": alertEvent.CreatedAt.Format(time.RFC3339),
		"BaseURL":     s.baseURL,
	}
}

func renderTemplateValue(value any, values map[string]any) any {
	switch typed := value.(type) {
	case map[string]any:
		rendered := make(map[string]any, len(typed))
		for key, nested := range typed {
			rendered[key] = renderTemplateValue(nested, values)
		}
		return rendered
	case []any:
		rendered := make([]any, len(typed))
		for i, nested := range typed {
			rendered[i] = renderTemplateValue(nested, values)
		}
		return rendered
	case string:
		return renderTemplateString(typed, values)
	default:
		return value
	}
}

func renderTemplateString(template string, values map[string]any) any {
	if key, ok := exactTemplateKey(template); ok {
		if value, exists := values[key]; exists {
			return value
		}
	}

	rendered := template
	for key, value := range values {
		rendered = strings.ReplaceAll(rendered, "{{."+key+"}}", fmt.Sprint(value))
	}
	return rendered
}

func exactTemplateKey(template string) (string, bool) {
	if !strings.HasPrefix(template, "{{.") || !strings.HasSuffix(template, "}}") {
		return "", false
	}

	key := strings.TrimSuffix(strings.TrimPrefix(template, "{{."), "}}")
	return key, key != ""
}

// logWebhookDelivery logs webhook delivery result to database
func (s *PushService) logWebhookDelivery(ctx context.Context, alertEvent *models.AlertEvent, webhook *models.Webhook, status string, retryCount int, errorMessage string) {
	log := &models.WebhookLog{
		WebhookID:    webhook.ID,
		AlertEventID: alertEvent.ID,
		Status:       status,
		RetryCount:   retryCount,
		ErrorMessage: errorMessage,
	}

	if err := s.webhookLogsQuerier.CreateWebhookLog(ctx, log); err != nil {
		slog.Error("Failed to create webhook log",
			"webhook_id", webhook.ID,
			"alert_id", alertEvent.ID,
			"status", status,
			"error", err)
	}
}
