package restHandler

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/golang-jwt/jwt/v5"
	muxContext "github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/iraunit/get-link-backend/pkg/services"
	"github.com/iraunit/get-link-backend/util/bean"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"sync"
)

type Links interface {
	SocketConnection(w http.ResponseWriter, r *http.Request)
	GetAllLinks(w http.ResponseWriter, r *http.Request)
	DeleteLinks(w http.ResponseWriter, r *http.Request)
	AddLink(w http.ResponseWriter, r *http.Request)
	VerifyWhatsappEmail(w http.ResponseWriter, r *http.Request)
}

type LinksImpl struct {
	logger      *zap.SugaredLogger
	Users       *map[string]bean.User
	lock        *sync.Mutex
	client      *redis.Client
	db          *pg.DB
	LinkService services.LinkService
	cfg         bean.MiddlewareCfg
}

func NewLinksImpl(logger *zap.SugaredLogger, client *redis.Client, db *pg.DB, users *map[string]bean.User, linkService services.LinkService) *LinksImpl {
	cfg := bean.MiddlewareCfg{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}
	return &LinksImpl{
		logger:      logger,
		Users:       users,
		lock:        &sync.Mutex{},
		client:      client,
		db:          db,
		LinkService: linkService,
		cfg:         cfg,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (impl *LinksImpl) SocketConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		impl.logger.Errorw("Error in upgrading connection to Web Sockets", "Error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 500, Error: "Error in upgrading connection to Web Sockets"})
		return
	}

	userEmail := muxContext.Get(r, "email").(string)
	impl.LinkService.HandleConnection(conn, userEmail)

	go impl.LinkService.ReadMessages(conn, userEmail, muxContext.Get(r, "uuid").(string))
	go impl.LinkService.WriteMessages(conn, userEmail)

}

func (impl *LinksImpl) GetAllLinks(w http.ResponseWriter, r *http.Request) {

	userEmail := muxContext.Get(r, "email").(string)
	uuid := muxContext.Get(r, "uuid").(string)
	links := impl.LinkService.GetAllLink(userEmail, uuid)
	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: links})
}

func (impl *LinksImpl) AddLink(w http.ResponseWriter, r *http.Request) {
	userEmail := muxContext.Get(r, "email").(string)

	var data bean.GetLink
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		impl.logger.Errorw("Error in decoding request body", "Error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error in decoding request body"})
		return
	}

	if data.UUID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "[uuid] is missing."})
		return
	}

	if data.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "[message] is missing."})
		return
	}

	impl.LinkService.AddLink(userEmail, &data)
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: "Link added successfully"})
}

func (impl *LinksImpl) DeleteLinks(w http.ResponseWriter, r *http.Request) {
	userEmail := muxContext.Get(r, "email").(string)
	var data bean.GetLink
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		impl.logger.Errorw("Error in decoding request body", "Error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 500, Error: "Error in decoding request body"})
		return
	}
	if data.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "[id] is missing."})
		return
	}
	err = impl.LinkService.DeleteLink(userEmail, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error in deleting link"})
		return
	}
	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: "Link deleted successfully"})
}

func (impl *LinksImpl) VerifyWhatsappEmail(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	tokenStr, _ := url.QueryUnescape(query.Get("token"))

	claims := bean.WhatsappVerificationClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(impl.cfg.JwtKey), nil
	})

	if err != nil {
		impl.logger.Errorw("Unauthorised Request. Invalid token.")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error parsing token."})
		return
	}

	err = impl.LinkService.VerifyWhatsapp(claims.Email, &bean.WhatsappEmail{Email: claims.Email, WhatAppNumber: claims.WhatAppNumber})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error in verifying link"})
		return
	}
	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: "Link verified successfully"})
}
