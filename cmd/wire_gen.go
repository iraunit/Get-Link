// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/iraunit/get-link-backend/api/restHandler"
	"github.com/iraunit/get-link-backend/api/router"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/util"
)

// Injectors from wire.go:

func InitializeApp() *App {
	sugaredLogger := util.InitLogger()
	middlewareImpl := router.NewMiddlewareImpl(sugaredLogger)
	client := repository.NewRedis(sugaredLogger)
	db := repository.NewPgDb(sugaredLogger)
	linksImpl := restHandler.NewLinksImpl(sugaredLogger, client, db)
	muxRouter := router.NewMuxRouter(middlewareImpl, linksImpl)
	app := NewApp(sugaredLogger, muxRouter)
	return app
}
