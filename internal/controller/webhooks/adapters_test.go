package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"testing"
)

// ---------------------------------------------------------------------------
// computeHMAC
// ---------------------------------------------------------------------------

func TestComputeHMAC(t *testing.T) {
	t.Run("deterministic", func(t *testing.T) {
		a := computeHMAC([]byte("payload"), "secret")
		b := computeHMAC([]byte("payload"), "secret")
		if a != b {
			t.Fatalf("expected identical HMAC for same inputs, got %q and %q", a, b)
		}
	})

	t.Run("different key produces different hmac", func(t *testing.T) {
		a := computeHMAC([]byte("payload"), "key-1")
		b := computeHMAC([]byte("payload"), "key-2")
		if a == b {
			t.Fatal("expected different HMAC for different keys")
		}
	})

	t.Run("different payload produces different hmac", func(t *testing.T) {
		a := computeHMAC([]byte("payload-a"), "secret")
		b := computeHMAC([]byte("payload-b"), "secret")
		if a == b {
			t.Fatal("expected different HMAC for different payloads")
		}
	})

	t.Run("valid 64 char hex string", func(t *testing.T) {
		result := computeHMAC([]byte("data"), "key")
		if len(result) != 64 {
			t.Fatalf("expected 64 hex chars, got %d", len(result))
		}
		if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(result) {
			t.Fatalf("result is not a valid lowercase hex string: %q", result)
		}
	})

	t.Run("known value", func(t *testing.T) {
		payload := []byte("test-payload")
		key := "test-secret"

		mac := hmac.New(sha256.New, []byte(key))
		mac.Write(payload)
		want := hex.EncodeToString(mac.Sum(nil))

		got := computeHMAC(payload, key)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

// ---------------------------------------------------------------------------
// discordColour
// ---------------------------------------------------------------------------

func TestDiscordColour(t *testing.T) {
	tests := []struct {
		eventType string
		want      int
	}{
		{"job.completed", 0x2ECC71},
		{"agent.online", 0x2ECC71},
		{"job.failed", 0xE74C3C},
		{"agent.offline", 0xE74C3C},
		{"job.cancelled", 0xF39C12},
		{"job.started", 0x3498DB},
		{"unknown.event", 0x3498DB},
		{"", 0x3498DB},
	}
	for _, tt := range tests {
		t.Run(tt.eventType, func(t *testing.T) {
			got := discordColour(tt.eventType)
			if got != tt.want {
				t.Fatalf("discordColour(%q) = 0x%X, want 0x%X", tt.eventType, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// payloadDescription
// ---------------------------------------------------------------------------

func TestPayloadDescription(t *testing.T) {
	t.Run("event with job_id", func(t *testing.T) {
		e := Event{
			Type:    "job.completed",
			Payload: map[string]any{"job_id": "abc-123"},
		}
		got := payloadDescription(e)
		want := "Job abc-123 \u2014 job.completed"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("event with agent_id", func(t *testing.T) {
		e := Event{
			Type:    "agent.online",
			Payload: map[string]any{"agent_id": "node-7"},
		}
		got := payloadDescription(e)
		want := "Agent node-7 \u2014 agent.online"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("event with neither", func(t *testing.T) {
		e := Event{
			Type:    "system.startup",
			Payload: map[string]any{"version": "1.0"},
		}
		got := payloadDescription(e)
		if got != "system.startup" {
			t.Fatalf("got %q, want %q", got, "system.startup")
		}
	})

	t.Run("job_id takes precedence over agent_id", func(t *testing.T) {
		e := Event{
			Type:    "job.completed",
			Payload: map[string]any{"job_id": "j-1", "agent_id": "a-1"},
		}
		got := payloadDescription(e)
		want := "Job j-1 \u2014 job.completed"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// formatPayload
// ---------------------------------------------------------------------------

func TestFormatPayload(t *testing.T) {
	event := Event{
		Type:    "job.completed",
		Payload: map[string]any{"job_id": "j-42", "status": "done"},
	}

	t.Run("discord", func(t *testing.T) {
		data, err := formatPayload("discord", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// username
		if parsed["username"] != "DistEncoder" {
			t.Fatalf("expected username DistEncoder, got %v", parsed["username"])
		}

		// embeds array
		embeds, ok := parsed["embeds"].([]any)
		if !ok || len(embeds) == 0 {
			t.Fatal("expected non-empty embeds array")
		}

		embed, ok := embeds[0].(map[string]any)
		if !ok {
			t.Fatal("expected embed to be an object")
		}

		// embed title matches event type
		if embed["title"] != event.Type {
			t.Fatalf("expected embed title %q, got %v", event.Type, embed["title"])
		}

		// embed colour is present
		if _, ok := embed["color"]; !ok {
			t.Fatal("expected embed to contain color field")
		}
	})

	t.Run("teams", func(t *testing.T) {
		data, err := formatPayload("teams", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		if parsed["type"] != "message" {
			t.Fatalf("expected type message, got %v", parsed["type"])
		}

		attachments, ok := parsed["attachments"].([]any)
		if !ok || len(attachments) == 0 {
			t.Fatal("expected non-empty attachments array")
		}

		att, ok := attachments[0].(map[string]any)
		if !ok {
			t.Fatal("expected attachment to be an object")
		}

		if att["contentType"] != "application/vnd.microsoft.card.adaptive" {
			t.Fatalf("unexpected contentType: %v", att["contentType"])
		}
	})

	t.Run("slack", func(t *testing.T) {
		data, err := formatPayload("slack", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		blocks, ok := parsed["blocks"].([]any)
		if !ok || len(blocks) == 0 {
			t.Fatal("expected non-empty blocks array")
		}

		first, ok := blocks[0].(map[string]any)
		if !ok {
			t.Fatal("expected first block to be an object")
		}

		if first["type"] != "header" {
			t.Fatalf("expected first block type header, got %v", first["type"])
		}
	})

	t.Run("generic unknown provider", func(t *testing.T) {
		data, err := formatPayload("custom", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// Should be a direct marshal of event.Payload, so keys match.
		if parsed["job_id"] != "j-42" {
			t.Fatalf("expected job_id j-42, got %v", parsed["job_id"])
		}
		if parsed["status"] != "done" {
			t.Fatalf("expected status done, got %v", parsed["status"])
		}

		// Should NOT contain provider-specific structure.
		if _, ok := parsed["username"]; ok {
			t.Fatal("generic payload should not contain discord username")
		}
		if _, ok := parsed["blocks"]; ok {
			t.Fatal("generic payload should not contain slack blocks")
		}
		if _, ok := parsed["attachments"]; ok {
			t.Fatal("generic payload should not contain teams attachments")
		}
	})
}
