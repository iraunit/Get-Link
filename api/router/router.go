package router

import (
	"github.com/gorilla/mux"
	"github.com/iraunit/get-link-backend/api/restHandler"
)

type MuxRouter struct {
	Router     *mux.Router
	middleware Middleware
	Links      restHandler.Links
}

func NewMuxRouter(middleware Middleware, links restHandler.Links) *MuxRouter {
	return &MuxRouter{
		Router:     mux.NewRouter(),
		middleware: middleware,
		Links:      links,
	}
}

func (r *MuxRouter) GetRouter() *mux.Router {
	r.Router.Use(r.middleware.CorsMiddleware)
	r.Router.Use(r.middleware.LoggerMiddleware)
	r.Router.Use(r.middleware.AuthMiddleware)

	r.Router.HandleFunc("/get-all-links", r.Links.GetAllLinks).Methods("GET")

	return r.Router
}
