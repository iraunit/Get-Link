package tokenService

import (
	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"time"
)

type TokenService interface {
	WhatsappEmailVerificationToken(claims *bean.WhatsappVerificationClaims) (string, error)
	TelegramEmailVerificationToken(claims *bean.TelegramVerificationClaims) (string, error)
	ShareFileVerificationToken(claims *bean.ShareFileClaims) (string, error)
}

type TokenServiceImpl struct {
	logger *zap.SugaredLogger
	cfg    *bean.TokenConfig
}

func NewTokenServiceImpl(logger *zap.SugaredLogger) *TokenServiceImpl {
	cfg := &bean.TokenConfig{}
	if err := env.Parse(cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}
	return &TokenServiceImpl{
		logger: logger,
		cfg:    cfg,
	}
}

func (impl *TokenServiceImpl) WhatsappEmailVerificationToken(claims *bean.WhatsappVerificationClaims) (string, error) {
	claims.ExpiresAt = &jwt.NumericDate{
		Time: time.Now().Add(24 * time.Hour),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(impl.cfg.JwtKey))
	return tokenStr, err
}

func (impl *TokenServiceImpl) TelegramEmailVerificationToken(claims *bean.TelegramVerificationClaims) (string, error) {
	claims.ExpiresAt = &jwt.NumericDate{
		Time: time.Now().Add(24 * time.Hour),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(impl.cfg.JwtKey))
	return tokenStr, err
}

func (impl *TokenServiceImpl) ShareFileVerificationToken(claims *bean.ShareFileClaims) (string, error) {
	claims.ExpiresAt = &jwt.NumericDate{
		Time: time.Now().Add(240 * time.Hour),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(impl.cfg.JwtKey))
	return tokenStr, err
}
