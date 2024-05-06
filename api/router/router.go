package router

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/iraunit/get-link-backend/api/restHandler"
	"github.com/iraunit/get-link-backend/util"
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
	r.Router.HandleFunc(fmt.Sprintf("/%s", util.WhatsappWebhook), r.Links.VerifyWhatsappEmail).Methods("GET")
	r.Router.HandleFunc(fmt.Sprintf("/%s", util.WHATSAPP), r.Whatsapp.Verify).Methods("GET")
	r.Router.HandleFunc(fmt.Sprintf("/%s", util.WHATSAPP), r.Whatsapp.HandleMessage).Methods("POST")

	return r.Router
}
