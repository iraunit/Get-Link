package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/iraunit/get-link-backend/api/router"
	"github.com/iraunit/get-link-backend/util"
	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		app.Logger.Errorw("Error loading Env", zap.Error(err))
	}

	cfg := util.MainCfg{}
	if err := env.Parse(&cfg); err != nil {
		app.Logger.Errorw("Error loading Cfg from env", zap.Error(err))
	}

	app.Logger.Infow(fmt.Sprintf("Starting server on port %s", cfg.Port))
	app.MuxRouter.GetRouter()

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), app.MuxRouter.Router); err != nil {
		app.Logger.Fatal(fmt.Sprintf("Cannot start server on port %s", cfg.Port), "Error: ", err)
	}
}
