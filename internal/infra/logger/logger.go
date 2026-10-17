package logger

import (
	"log/slog"

	"go.uber.org/fx"
)

var Provide = fx.Provide(New)

func New() (*slog.Logger, error) {
	return slog.Default(), nil
}
