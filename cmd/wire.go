//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/iraunit/get-link-backend/api/restHandler"
	"github.com/iraunit/get-link-backend/api/router"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/util"
)

func InitializeApp() *App {
	wire.Build(
		NewApp,
		util.InitLogger,
		router.NewMuxRouter,
		repository.NewRedis,
		repository.NewPgDb,
		router.NewMiddlewareImpl, wire.Bind(new(router.Middleware), new(*router.MiddlewareImpl)),
		restHandler.NewLinksImpl, wire.Bind(new(restHandler.Links), new(*restHandler.LinksImpl)),
	)
	return &App{}
}
