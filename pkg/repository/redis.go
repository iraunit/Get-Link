package repository

import (
	"github.com/caarlos0/env"
	"github.com/iraunit/get-link-backend/util/bean"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedis(logger *zap.SugaredLogger) *redis.Client {
	cfg := bean.RedisCfg{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading RedisCfg from env", "Error", zap.Error(err))
	}
	//return redis.NewClient(&redis.Options{
	//	Addr:     cfg.Addr,
	//	Password: cfg.Password,
	//	DB:       cfg.DB,
	//})

	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		logger.Fatal("Error parsing Redis URL", "Error", zap.Error(err))
	}

	return redis.NewClient(opt)
}
