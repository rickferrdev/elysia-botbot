package app

import (
	"github.com/rickferrdev/elysia-botbot/internal/config"
	"github.com/rickferrdev/elysia-botbot/internal/core/services"
	"github.com/rickferrdev/elysia-botbot/internal/inbound"
	"github.com/rickferrdev/elysia-botbot/internal/infra"
	"github.com/rickferrdev/elysia-botbot/internal/outbound"
	"go.uber.org/fx"
)

func Run() {
	fx.New(
		config.Module,
		infra.Module,
		outbound.Module,
		services.Module,
		inbound.Module,
	).Run()
}
