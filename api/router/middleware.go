package router

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/http"
)

type Middleware interface {
	AuthMiddleware(next http.Handler) http.Handler
	LoggerMiddleware(next http.Handler) http.Handler
}

type MiddlewareImpl struct {
	logger *zap.SugaredLogger
	cfg    bean.MiddlewareCfg
}

func NewMiddlewareImpl(logger *zap.SugaredLogger) *MiddlewareImpl {
	cfg := bean.MiddlewareCfg{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}

	return &MiddlewareImpl{
		logger: logger,
		cfg:    cfg,
	}
}

func (impl *MiddlewareImpl) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/whatsapp" || r.URL.Path == "/verifyWhatsapp" {
			if r.URL.Path == "/verifyEmail" {
				context.Set(r, "uuid", "Verify Email")
			} else if r.URL.Path == "/whatsapp" {
				context.Set(r, "uuid", "Whatsapp")
			}
			next.ServeHTTP(w, r)
			return
		} else {
			query := r.URL.Query()
			tokenStr := query.Get("Authorization")

			claims := bean.Claims{}
			_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(impl.cfg.JwtKey), nil
			})

			if err != nil {
				impl.logger.Errorw("Unauthorised Request. Invalid token.")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error parsing token."})
				return
			}

			device := query.Get("device")
			context.Set(r, "email", claims.Email)
			context.Set(r, "uuid", claims.UUID)
			context.Set(r, "device", device)
			next.ServeHTTP(w, r)
		}
	})
}

func (impl *MiddlewareImpl) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		impl.logger.Infow("req:", "Path:", r.URL.Path, "Method:", r.Method, "Email:", context.Get(r, "email"), "UUID:", context.Get(r, "uuid"))
		next.ServeHTTP(w, r)
	})
}
