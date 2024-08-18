package repository

import (
	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/iraunit/get-link-backend/util/bean"
)
import "go.uber.org/zap"

func NewPgDb(logger *zap.SugaredLogger) *pg.DB {
	cfg := bean.PgDbCfg{}

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
		"message" VARCHAR(102400),
		"sender" VARCHAR(64),
		"receiver" VARCHAR(64),
		"uuid" VARCHAR(512)
	  );`)

	if err != nil {
		logger.Fatal("Error creating schema for users", zap.Error(err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "whatsapp_emails" (
		"email" VARCHAR (512) PRIMARY KEY,
		"whatsapp_number" VARCHAR(512)
	  );`)

	if err != nil {
		logger.Fatal("Error creating schema for whatsapp_email", zap.Error(err))
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS "telegram_emails" (
		"email" VARCHAR (512) PRIMARY KEY,
		"sender_id" VARCHAR(512),
        "chat_id" VARCHAR(512)
	  );`)

	if err != nil {
		logger.Fatal("Error creating schema for telegram_email", zap.Error(err))
	}

	return db
}
