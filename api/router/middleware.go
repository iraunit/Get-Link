package router

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/http"
	"strings"
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
		if r.URL.Path == util.WhatsappWebhook || r.URL.Path == util.VerifyWhatsappEmail || r.URL.Path == util.VerifyTelegramEmail {
			if r.URL.Path == util.VerifyWhatsappEmail || r.URL.Path == util.VerifyTelegramEmail {
				context.Set(r, util.UUID, "Verify Email")
			} else if r.URL.Path == util.WhatsappWebhook {
				context.Set(r, util.UUID, util.WHATSAPP)
			}
			next.ServeHTTP(w, r)
			return
		} else if strings.Contains(r.URL.Path, "/download-shared-file/") {
			query := r.URL.Query()
			tokenStr := query.Get(util.AUTHORIZATION)

			claims := bean.ShareFileClaims{}
			_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(impl.cfg.JwtKey), nil
			})

			if err != nil {
				if tokenStr != "undefined" {
					impl.logger.Errorw("Unauthorised Request. Invalid token.", "URL", r.URL.Path, "Error", err)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error parsing token."})
				return
			}
			context.Set(r, util.EMAIL, claims.Email)
			next.ServeHTTP(w, r)
		} else {
			query := r.URL.Query()
			tokenStr := query.Get(util.AUTHORIZATION)

			claims := bean.Claims{}
			_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(impl.cfg.JwtKey), nil
			})

			if err != nil {
				if tokenStr != "undefined" {
					impl.logger.Errorw("Unauthorised Request. Invalid token.", "URL", r.URL.Path, "Error", err)
				}
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error parsing token."})
				return
			}

			device := query.Get(util.DEVICE)
			context.Set(r, util.EMAIL, claims.Email)
			context.Set(r, util.UUID, claims.UUID)
			context.Set(r, util.DEVICE, device)
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		}
	})
}

func (impl *MiddlewareImpl) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == util.WhatsappWebhook || r.URL.Path == util.VerifyWhatsappEmail {
			impl.logger.Infow("req:", "Path:", r.URL.Path, "Method:", r.Method, "UUID:", context.Get(r, "uuid"))
			next.ServeHTTP(w, r)
			return
		}
		impl.logger.Infow("req:", "Path:", r.URL.Path, "Method:", r.Method, "Email:", context.Get(r, "email"), "UUID:", context.Get(r, "uuid"))
		next.ServeHTTP(w, r)
	})
}
