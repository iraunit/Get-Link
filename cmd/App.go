package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/gorilla/handlers"
	"github.com/iraunit/get-link-backend/api/router"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/http"
)

type App struct {
	MuxRouter *router.MuxRouter
	Logger    *zap.SugaredLogger
}

func NewApp(logger *zap.SugaredLogger, muxRouter *router.MuxRouter) *App {
	return &App{
		MuxRouter: muxRouter,
		Logger:    logger,
	}
}

func (app *App) Start() {

	cfg := bean.MainCfg{}
	if err := env.Parse(&cfg); err != nil {
		app.Logger.Errorw("Error loading Cfg from env", zap.Error(err))
	}

	app.Logger.Infow(fmt.Sprintf("Starting server on port %s", cfg.Port))
	app.MuxRouter.GetRouter()

	corsOrigins := handlers.AllowedOrigins([]string{util.ChromeExtensionUrl, "https://getlink.codingkaro.in", "https://getlink.shyptsolution.com"})
	corsHeaders := handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding"})
	corsMethods := handlers.AllowedMethods([]string{"POST", "DELETE", "GET", "OPTIONS"})

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), handlers.CORS(corsOrigins, corsHeaders, corsMethods)(app.MuxRouter.Router)); err != nil {
		app.Logger.Fatal(fmt.Sprintf("Cannot start server on port %s", cfg.Port), "Error: ", err)
	}
}
