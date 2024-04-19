package util

import (
	"github.com/caarlos0/env"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"log"
)

func NewRedis() *redis.Client {
	cfg := RedisCfg{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal("Error loading RedisCfg from env", "Error", zap.Error(err))
	}
	//return redis.NewClient(&redis.Options{
	//	Addr:     cfg.Addr,
	//	Password: cfg.Password,
	//	DB:       cfg.DB,
	//})

	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		log.Fatal("Error parsing Redis URL", "Error", zap.Error(err))
	}

	return redis.NewClient(opt)
}
