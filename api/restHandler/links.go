package restHandler

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type Links interface {
	GetAllLinks(w http.ResponseWriter, r *http.Request)
}

type LinksImpl struct {
	logger *zap.SugaredLogger
}

func NewLinksImpl(logger *zap.SugaredLogger) *LinksImpl {
	return &LinksImpl{
		logger: logger,
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (impl *LinksImpl) GetAllLinks(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	// Simple echo server
	for {
		messageType, p, err := conn.ReadMessage()
		impl.logger.Infow("Info:", "messageType", messageType, "Message is : ", string(p), "err", err)
		if err != nil {
			log.Println(err)
			return
		}
		if err := conn.WriteMessage(messageType, []byte("Message from server")); err != nil {
			log.Println(err)
			return
		}
	}
}
