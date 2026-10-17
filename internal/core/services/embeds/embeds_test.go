package embeds_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/embeds"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
)

type mockEmbedsRepo struct {
	createOneFn func(ctx context.Context, embed domain.Embed) (*domain.Embed, error)
	upsertOneFn func(ctx context.Context, embed domain.Embed) (*domain.Embed, error)
	getOneFn    func(ctx context.Context, id int) (*domain.Embed, error)
	deleteOneFn func(ctx context.Context, id int) (*domain.Embed, error)
	getAllFn    func(ctx context.Context, guildID, authorID string) ([]domain.Embed, error)
}

func (m *mockEmbedsRepo) CreateOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
	if m.createOneFn != nil {
		return m.createOneFn(ctx, embed)
	}
	return &embed, nil
}

func (m *mockEmbedsRepo) UpsertOne(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
	if m.upsertOneFn != nil {
		return m.upsertOneFn(ctx, embed)
	}
	return &embed, nil
}

func (m *mockEmbedsRepo) DeleteOne(ctx context.Context, id int) (*domain.Embed, error) {
	if m.deleteOneFn != nil {
		return m.deleteOneFn(ctx, id)
	}
	return nil, nil
}

func (m *mockEmbedsRepo) GetOne(ctx context.Context, id int) (*domain.Embed, error) {
	if m.getOneFn != nil {
		return m.getOneFn(ctx, id)
	}
	return nil, nil
}

func (m *mockEmbedsRepo) GetAll(ctx context.Context, guildID, authorID string) ([]domain.Embed, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, guildID, authorID)
	}
	return nil, nil
}

func TestHandlerMessageContent(t *testing.T) {
	ctx := context.Background()
	validJSON := `{"content": "hello world", "embeds": []}`

	t.Run("invalid json returns ErrInvalidInput", func(t *testing.T) {
		svc := embeds.New(embeds.FxArgs{EmbedsDB: &mockEmbedsRepo{}})
		_, err := svc.HandlerMessageContent(ctx, 0, "guild1", "user1", "not-a-valid-json")
		if !errors.Is(err, embeds.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("id == 0 calls CreateOne", func(t *testing.T) {
		createCalled := false
		repo := &mockEmbedsRepo{
			createOneFn: func(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
				createCalled = true
				embed.ID = 1
				return &embed, nil
			},
		}
		svc := embeds.New(embeds.FxArgs{EmbedsDB: repo})
		res, err := svc.HandlerMessageContent(ctx, 0, "guild1", "user1", validJSON)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !createCalled {
			t.Error("expected CreateOne to be called for id == 0")
		}
		if res.ID != 1 {
			t.Errorf("expected ID 1, got %d", res.ID)
		}
	})

	t.Run("id > 0 calls UpsertOne", func(t *testing.T) {
		upsertCalled := false
		repo := &mockEmbedsRepo{
			upsertOneFn: func(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
				upsertCalled = true
				return &embed, nil
			},
		}
		svc := embeds.New(embeds.FxArgs{EmbedsDB: repo})
		res, err := svc.HandlerMessageContent(ctx, 10, "guild1", "user1", validJSON)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !upsertCalled {
			t.Error("expected UpsertOne to be called for id > 0")
		}
		if res.ID != 10 {
			t.Errorf("expected ID 10, got %d", res.ID)
		}
	})

	t.Run("conflict error returns ErrConflict", func(t *testing.T) {
		repo := &mockEmbedsRepo{
			createOneFn: func(ctx context.Context, embed domain.Embed) (*domain.Embed, error) {
				return nil, core.NewAppError("conflict", repoerrors.CodeSqlConflict, nil)
			},
		}
		svc := embeds.New(embeds.FxArgs{EmbedsDB: repo})
		_, err := svc.HandlerMessageContent(ctx, 0, "guild1", "user1", validJSON)
		if !errors.Is(err, embeds.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})
}

func TestObtainOne(t *testing.T) {
	ctx := context.Background()

	t.Run("not found mapping", func(t *testing.T) {
		repo := &mockEmbedsRepo{
			getOneFn: func(ctx context.Context, id int) (*domain.Embed, error) {
				return nil, core.NewAppError("not found", repoerrors.CodeSqlNoRows, nil)
			},
		}
		svc := embeds.New(embeds.FxArgs{EmbedsDB: repo})
		_, err := svc.ObtainOne(ctx, 999)
		if !errors.Is(err, embeds.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("successful fetch", func(t *testing.T) {
		repo := &mockEmbedsRepo{
			getOneFn: func(ctx context.Context, id int) (*domain.Embed, error) {
				return &domain.Embed{ID: id, Content: "sample"}, nil
			},
		}
		svc := embeds.New(embeds.FxArgs{EmbedsDB: repo})
		res, err := svc.ObtainOne(ctx, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.ID != 5 {
			t.Errorf("expected ID 5, got %d", res.ID)
		}
	})
}
