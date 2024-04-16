package router

import (
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
}

func NewMiddlewareImpl(logger *zap.SugaredLogger) *MiddlewareImpl {
	return &MiddlewareImpl{
		logger: logger,
	}
}

func (impl *MiddlewareImpl) CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (impl *MiddlewareImpl) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func (impl *MiddlewareImpl) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		impl.logger.Infow("req:", "Path:", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
