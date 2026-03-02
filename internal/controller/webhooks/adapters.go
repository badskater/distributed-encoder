package webhooks

import (
	"encoding/json"
	"fmt"
)

// formatPayload converts an Event into the provider-specific JSON payload.
func formatPayload(provider string, event Event) ([]byte, error) {
	switch provider {
	case "discord":
		return formatDiscord(event)
	case "teams":
		return formatTeams(event)
	case "slack":
		return formatSlack(event)
	default:
		return json.Marshal(event.Payload)
	}
}

// formatDiscord produces a Discord webhook payload using embeds.
func formatDiscord(event Event) ([]byte, error) {
	colour := discordColour(event.Type)
	embed := map[string]any{
		"title":       event.Type,
		"color":       colour,
		"description": payloadDescription(event),
		"fields":      payloadFields(event),
	}
	payload := map[string]any{
		"username": "DistEncoder",
		"embeds":   []any{embed},
	}
	return json.Marshal(payload)
}

// formatTeams produces a Microsoft Teams Adaptive Card payload.
func formatTeams(event Event) ([]byte, error) {
	body := []map[string]any{
		{"type": "TextBlock", "text": event.Type, "weight": "Bolder", "size": "Medium"},
		{"type": "TextBlock", "text": payloadDescription(event), "wrap": true},
	}
	for k, v := range event.Payload {
		body = append(body, map[string]any{
			"type":     "TextBlock",
			"text":     fmt.Sprintf("**%s**: %v", k, v),
			"wrap":     true,
			"isSubtle": true,
		})
	}
	payload := map[string]any{
		"type": "message",
		"attachments": []any{
			map[string]any{
				"contentType": "application/vnd.microsoft.card.adaptive",
				"content": map[string]any{
					"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
					"type":    "AdaptiveCard",
					"version": "1.4",
					"body":    body,
				},
			},
		},
	}
	return json.Marshal(payload)
}

// formatSlack produces a Slack Block Kit payload.
func formatSlack(event Event) ([]byte, error) {
	blocks := []map[string]any{
		{
			"type": "header",
			"text": map[string]any{"type": "plain_text", "text": event.Type},
		},
		{
			"type": "section",
			"text": map[string]any{"type": "mrkdwn", "text": payloadDescription(event)},
		},
	}
	for k, v := range event.Payload {
		blocks = append(blocks, map[string]any{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": fmt.Sprintf("*%s*: %v", k, v),
			},
		})
	}
	payload := map[string]any{"blocks": blocks}
	return json.Marshal(payload)
}

// discordColour returns a decimal colour code based on event type.
func discordColour(eventType string) int {
	switch {
	case eventType == "job.completed" || eventType == "agent.online":
		return 0x2ECC71 // green
	case eventType == "job.failed" || eventType == "agent.offline":
		return 0xE74C3C // red
	case eventType == "job.cancelled":
		return 0xF39C12 // orange
	default:
		return 0x3498DB // blue
	}
}

// payloadDescription extracts a summary from the event payload.
func payloadDescription(event Event) string {
	if id, ok := event.Payload["job_id"].(string); ok {
		return fmt.Sprintf("Job %s — %s", id, event.Type)
	}
	if id, ok := event.Payload["agent_id"].(string); ok {
		return fmt.Sprintf("Agent %s — %s", id, event.Type)
	}
	return event.Type
}

// payloadFields returns the event payload as a slice of Discord embed field maps.
func payloadFields(event Event) []map[string]any {
	fields := make([]map[string]any, 0, len(event.Payload))
	for k, v := range event.Payload {
		fields = append(fields, map[string]any{
			"name":   k,
			"value":  fmt.Sprintf("%v", v),
			"inline": true,
		})
	}
	return fields
}
