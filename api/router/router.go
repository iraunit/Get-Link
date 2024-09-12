package router

import (
	"github.com/gorilla/mux"
	"github.com/iraunit/get-link-backend/api/restHandler"
)

type MuxRouter struct {
	Router      *mux.Router
	middleware  Middleware
	Links       restHandler.Links
	Whatsapp    restHandler.Whatsapp
	fileHandler restHandler.FileHandler
	Telegram    restHandler.TelegramRestHandler
}

func NewMuxRouter(middleware Middleware, links restHandler.Links, whatsapp restHandler.Whatsapp, fileHandler restHandler.FileHandler, telegram restHandler.TelegramRestHandler) *MuxRouter {
	return &MuxRouter{
		Router:      mux.NewRouter(),
		middleware:  middleware,
		Links:       links,
		Whatsapp:    whatsapp,
		fileHandler: fileHandler,
		Telegram:    telegram,
	}
}

func (r *MuxRouter) GetRouter() *mux.Router {

	r.Router.Use(r.middleware.AuthMiddleware)
	r.Router.Use(r.middleware.LoggerMiddleware)

	r.Router.HandleFunc("/", r.Links.AddLink).Methods("POST")
	r.Router.HandleFunc("/", r.Links.DeleteLinks).Methods("DELETE")
	r.Router.HandleFunc("/", r.Links.GetAllLinks).Methods("GET")
	r.Router.HandleFunc("/ws", r.Links.SocketConnection).Methods("GET")
	r.Router.HandleFunc("/verify-whatsapp-email", r.Links.VerifyWhatsappEmail).Methods("GET")
	r.Router.HandleFunc("/whatsapp-webhook", r.Whatsapp.Verify).Methods("GET")
	r.Router.HandleFunc("/whatsapp-webhook", r.Whatsapp.HandleMessage).Methods("POST")
	r.Router.HandleFunc("/download-file/{appName}/{fileName}", r.fileHandler.DownloadFile).Methods("GET")
	r.Router.HandleFunc("/download-shared-file/{appName}/{fileName}", r.fileHandler.DownloadFile).Methods("GET")
	r.Router.HandleFunc("/upload-file", r.fileHandler.UploadFile).Methods("POST")
	r.Router.HandleFunc("/list-files", r.fileHandler.ListAllFiles).Methods("GET")
	r.Router.HandleFunc("/delete-file/{appName}/{fileName}", r.fileHandler.DeleteFile).Methods("DELETE")
	r.Router.HandleFunc("/verify-telegram-email", r.Telegram.VerifyTelegramEmail).Methods("GET")
	r.Router.HandleFunc("/send-telegram-message", r.Telegram.SendTelegramMessage).Methods("POST")
	return r.Router
}
