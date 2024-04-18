package router

import (
	"github.com/caarlos0/env"
	"github.com/gorilla/context"
	"github.com/iraunit/get-link-backend/util"
	"go.uber.org/zap"
	"net/http"
)

type Middleware interface {
	CorsMiddleware(next http.Handler) http.Handler
	AuthMiddleware(next http.Handler) http.Handler
	LoggerMiddleware(next http.Handler) http.Handler
}

type MiddlewareImpl struct {
	logger *zap.SugaredLogger
	cfg    util.MiddlewareCfg
}

func NewMiddlewareImpl(logger *zap.SugaredLogger) *MiddlewareImpl {
	cfg := util.MiddlewareCfg{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}

	return &MiddlewareImpl{
		logger: logger,
		cfg:    cfg,
	}
}

func (impl *MiddlewareImpl) CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (impl *MiddlewareImpl) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//cookie, err := r.Cookie("Authorization")
		//if err != nil {
		//	impl.logger.Errorw("Unauthorised Request. No cookie found.")
		//	w.WriteHeader(http.StatusUnauthorized)
		//	_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 401, Error: "Unauthorised Request. No cookie found."})
		//	return
		//}
		//
		//tokenStr := cookie.Value
		//claims := util.Claims{}
		//_, err = jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		//	return []byte(impl.cfg.JwtKey), nil
		//})
		//
		//if err != nil {
		//	impl.logger.Errorw("Unauthorised Request. Invalid token.")
		//	w.WriteHeader(http.StatusBadRequest)
		//	_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 400, Error: "Error parsing token."})
		//	return
		//}

		context.Set(r, "email", "ramverma@gmail.com")
		next.ServeHTTP(w, r)
	})
}

func (impl *MiddlewareImpl) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		impl.logger.Infow("req:", "Path:", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
