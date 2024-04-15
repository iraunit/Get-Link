package router

import (
	"github.com/gorilla/mux"
	"github.com/iraunit/get-link-backend/api/restHandler"
)

type MuxRouter struct {
	Router *mux.Router
}

func NewMuxRouter() *MuxRouter {
	return &MuxRouter{
		Router: mux.NewRouter(),
	}
}

func (r *MuxRouter) GetRouter() *mux.Router {
	r.Router.Use(CorsMiddleware)
	r.Router.Use(AuthMiddleware)
	r.Router.Use(LoggerMiddleware)

	r.Router.HandleFunc("/get-all-links", restHandler.GetAllLinks).Methods("GET")

	return r.Router
}
