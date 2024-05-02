package restHandler

import (
	"encoding/json"
	"github.com/go-pg/pg"
	muxContext "github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/iraunit/get-link-backend/pkg/services"
	"github.com/iraunit/get-link-backend/util"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type Links interface {
	SocketConnection(w http.ResponseWriter, r *http.Request)
	GetAllLinks(w http.ResponseWriter, r *http.Request)
	DeleteLinks(w http.ResponseWriter, r *http.Request)
	AddLink(w http.ResponseWriter, r *http.Request)
}

type LinksImpl struct {
	logger      *zap.SugaredLogger
	Users       *map[string]util.User
	lock        *sync.Mutex
	client      *redis.Client
	db          *pg.DB
	LinkService services.LinkService
}

func NewLinksImpl(logger *zap.SugaredLogger, client *redis.Client, db *pg.DB, users *map[string]util.User, linkService services.LinkService) *LinksImpl {
	return &LinksImpl{
		logger:      logger,
		Users:       users,
		lock:        &sync.Mutex{},
		client:      client,
		db:          db,
		LinkService: linkService,
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
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 500, Error: "Error in upgrading connection to Web Sockets"})
		return
	}

	userEmail := muxContext.Get(r, "email").(string)
	impl.LinkService.HandleConnection(conn, userEmail)

	go impl.LinkService.ReadMessages(conn, userEmail, muxContext.Get(r, "uuid").(string), muxContext.Get(r, "device").(string))
	go impl.LinkService.WriteMessages(conn, userEmail, muxContext.Get(r, "uuid").(string), muxContext.Get(r, "device").(string))

}

func (impl *LinksImpl) GetAllLinks(w http.ResponseWriter, r *http.Request) {

	userEmail := muxContext.Get(r, "email").(string)
	uuid := muxContext.Get(r, "uuid").(string)
	device := muxContext.Get(r, "device").(string)
	links := impl.LinkService.GetAllLink(userEmail, uuid, device)
	_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 200, Result: links})
}

func (impl *LinksImpl) AddLink(w http.ResponseWriter, r *http.Request) {
	userEmail := muxContext.Get(r, "email").(string)

	var data util.GetLink
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		impl.logger.Errorw("Error in decoding request body", "Error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 400, Error: "Error in decoding request body"})
		return
	}

	if data.UUID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 400, Error: "[uuid] is missing."})
		return
	}

	if data.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 400, Error: "[message] is missing."})
		return
	}

	impl.LinkService.AddLink(userEmail, &data)
	_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 200, Result: "Link added successfully"})
}

func (impl *LinksImpl) DeleteLinks(w http.ResponseWriter, r *http.Request) {
	userEmail := muxContext.Get(r, "email").(string)
	var data util.GetLink
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		impl.logger.Errorw("Error in decoding request body", "Error: ", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 500, Error: "Error in decoding request body"})
		return
	}
	if data.ID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 400, Error: "[id] is missing."})
		return
	}
	err = impl.LinkService.DeleteLink(userEmail, &data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 400, Error: "Error in deleting link"})
		return
	}
	_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 200, Result: "Link deleted successfully"})
}
