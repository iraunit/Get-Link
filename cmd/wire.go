//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/iraunit/get-link-backend/api/restHandler"
	"github.com/iraunit/get-link-backend/api/router"
	"github.com/iraunit/get-link-backend/pkg/fileManager"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/pkg/restCalls"
	"github.com/iraunit/get-link-backend/pkg/services"
	"github.com/iraunit/get-link-backend/util"
)

func InitializeApp() *App {
	wire.Build(
		NewApp,
		util.InitLogger,
		router.NewMuxRouter,
		repository.NewUsersMap,
		repository.NewRedis,
		repository.NewPgDb,
		services.NewLinkServiceImpl, wire.Bind(new(services.LinkService), new(*services.LinkServiceImpl)),
		repository.NewRepositoryImpl, wire.Bind(new(repository.Repository), new(*repository.Impl)),
		router.NewMiddlewareImpl, wire.Bind(new(router.Middleware), new(*router.MiddlewareImpl)),
		restHandler.NewLinksImpl, wire.Bind(new(restHandler.Links), new(*restHandler.LinksImpl)),
		restHandler.NewWhatsappImpl, wire.Bind(new(restHandler.Whatsapp), new(*restHandler.WhatsappImpl)),
		services.NewWhatsappServiceImpl, wire.Bind(new(services.WhatsappService), new(*services.WhatsappServiceImpl)),
		restCalls.NewRestClientImpl, wire.Bind(new(restCalls.RestClient), new(*restCalls.RestClientImpl)),
		services.NewTokenServiceImpl, wire.Bind(new(services.TokenService), new(*services.TokenServiceImpl)),
		services.NewMailServiceImpl, wire.Bind(new(services.MailService), new(*services.MailServiceImpl)),
		fileManager.NewFileManagerImpl, wire.Bind(new(fileManager.FileManager), new(*fileManager.FileManagerImpl)),
		util.NewAsync,
		restHandler.NewFileHandlerImpl, wire.Bind(new(restHandler.FileHandler), new(*restHandler.FileHandlerImpl)),
		services.NewFileServiceImpl, wire.Bind(new(services.FileService), new(*services.FileServiceImpl)),
		restHandler.NewTelegramRestHandler, wire.Bind(new(restHandler.TelegramRestHandler), new(*restHandler.TelegramRestHandlerImpl)),
		services.NewTelegramService, wire.Bind(new(services.TelegramService), new(*services.TelegramImpl)),
	)
	return &App{}
}
