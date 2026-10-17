package server_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core"
	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/server"
	"github.com/rickferrdev/elysia-botbot/internal/outbound/databases/repoerrors"
)

type mockServerStorage struct {
	createFn func(ctx context.Context, s domain.Server) error
	getFn    func(ctx context.Context, guildID string) (*domain.Server, error)
}

func (m *mockServerStorage) CreateServerWithConfig(ctx context.Context, s domain.Server) error {
	if m.createFn != nil {
		return m.createFn(ctx, s)
	}
	return nil
}

func (m *mockServerStorage) GetServerWithConfig(ctx context.Context, guildID string) (*domain.Server, error) {
	if m.getFn != nil {
		return m.getFn(ctx, guildID)
	}
	return &domain.Server{GuildID: guildID}, nil
}

func TestServerService(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateServerWithConfig conflict mapping", func(t *testing.T) {
		storage := &mockServerStorage{
			createFn: func(ctx context.Context, s domain.Server) error {
				return core.NewAppError("conflict", repoerrors.CodeSqlConflict, nil)
			},
		}
		svc := server.New(server.FxArgs{ServerStorage: storage})
		err := svc.CreateServerWithConfig(ctx, domain.Server{GuildID: "123"})
		if !errors.Is(err, server.ErrConflict) {
			t.Errorf("expected ErrConflict, got %v", err)
		}
	})

	t.Run("ObtainServerWithConfig not found mapping", func(t *testing.T) {
		storage := &mockServerStorage{
			getFn: func(ctx context.Context, guildID string) (*domain.Server, error) {
				return nil, core.NewAppError("not found", repoerrors.CodeSqlNoRows, nil)
			},
		}
		svc := server.New(server.FxArgs{ServerStorage: storage})
		_, err := svc.ObtainServerWithConfig(ctx, "999")
		if !errors.Is(err, server.ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("ObtainServerWithConfig success", func(t *testing.T) {
		storage := &mockServerStorage{
			getFn: func(ctx context.Context, guildID string) (*domain.Server, error) {
				return &domain.Server{GuildID: guildID}, nil
			},
		}
		svc := server.New(server.FxArgs{ServerStorage: storage})
		s, err := svc.ObtainServerWithConfig(ctx, "abc")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.GuildID != "abc" {
			t.Errorf("expected GuildID 'abc', got %q", s.GuildID)
		}
	})
}
