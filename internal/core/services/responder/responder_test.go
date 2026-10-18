package responder_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/discord-go/discord.go/interactions"
	"github.com/discord-go/discord.go/messages"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/responder"
)

func TestResponderPresets(t *testing.T) {
	presets := []struct {
		name string
		data interactions.InteractionCallbackData
	}{
		{"WithOk", responder.WithOk},
		{"WithCreated", responder.WithCreated},
		{"WithNoContent", responder.WithNoContent},
		{"WithBadArguments", responder.WithBadArguments},
		{"WithUnauthorized", responder.WithUnauthorized},
		{"WithForbidden", responder.WithForbidden},
		{"WithNotFound", responder.WithNotFound},
		{"WithConflict", responder.WithConflict},
		{"WithInternal", responder.WithInternal},
		{"WithServiceUnavailable", responder.WithServiceUnavailable},
	}

	for _, tc := range presets {
		t.Run(tc.name, func(t *testing.T) {
			if tc.data.Content == "" {
				t.Errorf("%s: content should not be empty", tc.name)
			}
			if len(tc.data.Embeds) == 0 {
				t.Fatalf("%s: expected at least one embed", tc.name)
			}
			embed := tc.data.Embeds[0]
			if embed.Title == "" {
				t.Errorf("%s: embed title should not be empty", tc.name)
			}
			if embed.Description == "" {
				t.Errorf("%s: embed description should not be empty", tc.name)
			}
		})
	}
}

func TestWithEmbed(t *testing.T) {
	customEmbed := messages.Embed{
		Title:       "Custom Title",
		Description: "Custom Description",
	}

	result := responder.WithEmbed(customEmbed)
	if result.Content == "" {
		t.Error("expected non-empty content")
	}
	if len(result.Embeds) != 1 {
		t.Fatalf("expected 1 embed, got %d", len(result.Embeds))
	}
	if result.Embeds[0].Title != customEmbed.Title {
		t.Errorf("expected title %q, got %q", customEmbed.Title, result.Embeds[0].Title)
	}
}

func TestSender_NilContext(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := responder.New(responder.FxArgs{Logger: discardLogger})

	service.Sender(nil, responder.WithOk, nil)
	service.Sender(nil, responder.WithInternal, errors.New("database failure"))
	service.SenderContent(nil, responder.WithCodeContent(responder.CodeInternal), errors.New("db"))
}

func TestWithCodeContent(t *testing.T) {
	got := responder.WithCodeContent(responder.CodeInternal)
	if got.Title == "" || got.Description == "" {
		t.Fatal("expected title and description from known code")
	}
	fallback := responder.WithCodeContent(responder.ResponderCode(999))
	internal := responder.WithCodeContent(responder.CodeInternal)
	if fallback.Title != internal.Title {
		t.Fatalf("unknown code should fall back to internal, got %q", fallback.Title)
	}
}

func TestWithContent(t *testing.T) {
	got := responder.WithContent("hi", "there")
	if got.Title != "hi" || got.Description != "there" {
		t.Fatalf("unexpected content: %+v", got)
	}
}
