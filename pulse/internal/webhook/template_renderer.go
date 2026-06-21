package webhook

import (
	"fmt"
	"strings"
	"time"

	"github.com/whg517/node-pulse/pulse/internal/models"
)

// RenderAlertEvent renders a webhook event format with alert template values.
func RenderAlertEvent(baseURL string, alertEvent *models.AlertEvent, eventFormat map[string]any) (map[string]any, error) {
	if len(eventFormat) == 0 {
		eventFormat = models.DefaultEventFormat
	}

	formatted, ok := renderTemplateValue(eventFormat, alertTemplateValues(baseURL, alertEvent)).(map[string]any)
	if !ok {
		return nil, fmt.Errorf("webhook event format must be a JSON object")
	}
	return formatted, nil
}

func alertTemplateValues(baseURL string, alertEvent *models.AlertEvent) map[string]any {
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
		"BaseURL":     baseURL,
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
