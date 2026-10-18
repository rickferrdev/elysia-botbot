package funslipper_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core/domain"
	"github.com/rickferrdev/elysia-botbot/internal/core/services/funslipper"
)

type mockRepo struct {
	fn func(ctx context.Context, item domain.FunSlipper) (*domain.FunSlipper, error)
}

func (m *mockRepo) IncrementCount(ctx context.Context, item domain.FunSlipper) (*domain.FunSlipper, error) {
	return m.fn(ctx, item)
}

func TestIncrementCount(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		svc := funslipper.New(funslipper.FxArgs{
			FunSlipper: &mockRepo{fn: func(ctx context.Context, item domain.FunSlipper) (*domain.FunSlipper, error) {
				if item.GuildID != "g" || item.UserID != "u" {
					t.Fatalf("unexpected ids: %+v", item)
				}
				item.Count = 3
				return &item, nil
			}},
		})
		got, err := svc.IncrementCount(ctx, "g", "u")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Count != 3 {
			t.Fatalf("expected count 3, got %d", got.Count)
		}
	})

	t.Run("maps repo error", func(t *testing.T) {
		svc := funslipper.New(funslipper.FxArgs{
			FunSlipper: &mockRepo{fn: func(ctx context.Context, item domain.FunSlipper) (*domain.FunSlipper, error) {
				return nil, errors.New("db down")
			}},
		})
		_, err := svc.IncrementCount(ctx, "g", "u")
		if !errors.Is(err, funslipper.ErrConflict) {
			t.Fatalf("expected ErrConflict, got %v", err)
		}
	})
}
