package repository

import (
	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/iraunit/get-link-backend/util"
)
import "go.uber.org/zap"

func NewPgDb(logger *zap.SugaredLogger) *pg.DB {
	cfg := util.PgDbCfg{}

	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading PgDb cfg.", "Error:", err)
		return nil
	}

	return pg.Connect(&pg.Options{
		Addr:     cfg.Address,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
	})
}
