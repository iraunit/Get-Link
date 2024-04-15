package cmd

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	MuxRouter *mux.Router
	Logger    *zap.SugaredLogger
}

func NewApp(logger *zap.SugaredLogger, muxRouter *mux.Router) *App {
	return &App{
		MuxRouter: muxRouter,
		Logger:    logger,
	}
}

func (app *App) Start() error {
	app.Logger.Info("Starting server")

	return nil
}
