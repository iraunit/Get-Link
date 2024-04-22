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

	db := pg.Connect(&pg.Options{
		Addr:     cfg.Address,
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
	})

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS "get_links" (
		"id" SERIAL PRIMARY KEY,
		"destination" VARCHAR(512),
		"message" VARCHAR(102400),
		"uuid" VARCHAR(512)
	  );`)

	if err != nil {
		logger.Fatal("Error creating schema for users", zap.Error(err))
	}

	return db
}
