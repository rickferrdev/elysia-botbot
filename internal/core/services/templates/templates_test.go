package templates_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/templates"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repositories/dbtemplates/mocks"
	"go.uber.org/mock/gomock"
)

func TestHandlerMessageContentCreatesNewTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mock_templates.NewMockInterface(ctrl)
	repository.EXPECT().
		CreateOne(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, template domain.Template) (*domain.Template, error) {
			if template.TemplateContent.Content != "hello" {
				t.Errorf("unexpected content: %q", template.TemplateContent.Content)
			}

			var savedEmbeds []map[string]any
			if err := json.Unmarshal([]byte(template.TemplateContent.Embeds), &savedEmbeds); err != nil {
				t.Fatalf("decode saved embeds: %v", err)
			}
			if len(savedEmbeds) != 1 {
				t.Errorf("expected one saved embed, got %d", len(savedEmbeds))
			}
			template.ID = 1
			return &template, nil
		})

	service := templates.New(templates.FxArgs{TemplatesDB: repository})
	result, err := service.HandlerMessageContent(
		context.Background(),
		0,
		"guild",
		"author",
		`{"content":"hello","embeds":[{"title":"test"}],"attachments":[]}`,
	)
	if err != nil {
		t.Fatalf("HandlerMessageContent returned error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected created template ID 1, got %d", result.ID)
	}
}

func TestHandlerMessageContentUpdatesExistingTemplate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mock_templates.NewMockInterface(ctrl)
	repository.EXPECT().
		UpdateOne(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, template domain.Template) (*domain.Template, error) {
			if template.ID != 7 {
				t.Errorf("expected template ID 7, got %d", template.ID)
			}
			if template.TemplateContent.Content != "updated" {
				t.Errorf("unexpected content: %q", template.TemplateContent.Content)
			}
			return &template, nil
		})

	service := templates.New(templates.FxArgs{TemplatesDB: repository})
	result, err := service.HandlerMessageContent(
		context.Background(),
		7,
		"guild",
		"author",
		`{"content":"updated","embeds":[],"attachments":[]}`,
	)
	if err != nil {
		t.Fatalf("HandlerMessageContent returned error: %v", err)
	}
	if result.ID != 7 {
		t.Errorf("expected updated template ID 7, got %d", result.ID)
	}
}
