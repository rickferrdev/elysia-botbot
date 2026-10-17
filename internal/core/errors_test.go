package core_test

import (
	"errors"
	"testing"

	"github.com/rickferrdev/elysia-botbot/internal/core"
)

func TestAppError(t *testing.T) {
	cause := errors.New("timeout connection")
	err := core.NewAppError("operation failed", "ERR_CODE", cause)

	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}

	if !errors.Is(err.Unwrap(), cause) {
		t.Errorf("expected Unwrap to return cause")
	}

	sameCodeTarget := core.NewAppError("another message", "ERR_CODE", nil)
	if !err.Is(sameCodeTarget) {
		t.Errorf("expected Is to match error with same code")
	}

	differentCodeTarget := core.NewAppError("another message", "OTHER_CODE", nil)
	if err.Is(differentCodeTarget) {
		t.Errorf("expected Is to not match error with different code")
	}

	if !core.IsCode(err, "ERR_CODE") {
		t.Errorf("expected IsCode to return true for ERR_CODE")
	}

	if core.IsCode(err, "OTHER_CODE") {
		t.Errorf("expected IsCode to return false for OTHER_CODE")
	}

	if core.IsCode(errors.New("regular error"), "ERR_CODE") {
		t.Errorf("expected IsCode to return false for non-AppError")
	}

	errNoCause := core.NewAppError("no cause", "ERR_CODE", nil)
	if errNoCause.Error() == "" {
		t.Errorf("expected non-empty error message when cause is nil")
	}
}
