package router

import (
	"github.com/gorilla/mux"
	"github.com/iraunit/get-link-backend/api/restHandler"
)

type MuxRouter struct {
	Router     *mux.Router
	middleware Middleware
	Links      restHandler.Links
	Whatsapp   restHandler.Whatsapp
}

func NewMuxRouter(middleware Middleware, links restHandler.Links, whatsapp restHandler.Whatsapp) *MuxRouter {
	return &MuxRouter{
		Router:     mux.NewRouter(),
		middleware: middleware,
		Links:      links,
		Whatsapp:   whatsapp,
	}
}

func (r *MuxRouter) GetRouter() *mux.Router {

	r.Router.Use(r.middleware.AuthMiddleware)
	r.Router.Use(r.middleware.LoggerMiddleware)

	r.Router.HandleFunc("/", r.Links.AddLink).Methods("POST")
	r.Router.HandleFunc("/", r.Links.DeleteLinks).Methods("DELETE")
	r.Router.HandleFunc("/", r.Links.GetAllLinks).Methods("GET")
	r.Router.HandleFunc("/ws", r.Links.SocketConnection).Methods("GET")
	r.Router.HandleFunc("/verifyWhatsapp", r.Links.VerifyWhatsappNumber).Methods("GET")
	r.Router.HandleFunc("/whatsapp", r.Whatsapp.Verify).Methods("GET")
	r.Router.HandleFunc("/whatsapp", r.Whatsapp.HandleMessage).Methods("POST")

	return r.Router
}
