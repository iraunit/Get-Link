package restHandler

import (
	"context"
	"encoding/json"
	"fmt"
	muxContext "github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/iraunit/get-link-backend/util"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type Links interface {
	GetAllLinks(w http.ResponseWriter, r *http.Request)
	HandleDisconnection(conn *websocket.Conn, userEmail string)
	HandleConnection(conn *websocket.Conn, userEmail string)
	ReadMessages(conn *websocket.Conn, userEmail string)
	WriteMessages(conn *websocket.Conn, userEmail string)
}

type LinksImpl struct {
	logger *zap.SugaredLogger
	Users  map[string]util.User
	lock   *sync.Mutex
	client *redis.Client
}

func NewLinksImpl(logger *zap.SugaredLogger, client *redis.Client) *LinksImpl {
	return &LinksImpl{
		logger: logger,
		Users:  make(map[string]util.User),
		lock:   &sync.Mutex{},
		client: client,
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
		impl.logger.Errorw("Error in upgrading connection to Web Sockets", "Error: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(util.Response{StatusCode: 500, Error: "Error in upgrading connection to Web Sockets"})
		return
	}

	userEmail := muxContext.Get(r, "email").(string)
	impl.HandleConnection(conn, userEmail)

	//for {
	//	messageType, p, err := conn.ReadMessage()
	//	impl.logger.Infow("Info:", "messageType", messageType, "Message is : ", string(p), "err", err)
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//	if err := conn.WriteMessage(messageType, []byte("Message from server")); err != nil {
	//		log.Println(err)
	//		return
	//	}
	//}

	go impl.ReadMessages(conn, userEmail)
	go impl.WriteMessages(conn, userEmail)

}

func (impl *LinksImpl) HandleConnection(conn *websocket.Conn, userEmail string) {
	impl.lock.Lock()
	user, ok := impl.Users[userEmail]
	if !ok {
		user = util.User{
			Lock:        &sync.Mutex{},
			Connections: make([]*websocket.Conn, 0),
		}
		impl.Users[userEmail] = user
	}
	user.Lock.Lock()
	var exists bool
	for _, c := range user.Connections {
		if c == conn {
			exists = true
			break
		}
	}
	if !exists {
		user.Connections = append(user.Connections, conn)
	}
	user.Lock.Unlock()
	impl.lock.Unlock()
}

func (impl *LinksImpl) HandleDisconnection(conn *websocket.Conn, userEmail string) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	user, ok := impl.Users[userEmail]

	if !ok {
		impl.logger.Errorw("User not found in map.", "Email", userEmail)
		return
	}

	user.Lock.Lock()
	defer user.Lock.Unlock()

	index := -1
	for i, c := range user.Connections {
		if c == conn {
			index = i
			break
		}
	}

	if index != -1 {
		user.Connections = append(user.Connections[:index], user.Connections[index+1:]...)
	}

	if len(user.Connections) == 0 {
		delete(impl.Users, userEmail)
	}
}

func (impl *LinksImpl) ReadMessages(conn *websocket.Conn, userEmail string) {
	//Read message from Client and push to Redis

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			impl.logger.Errorw("Error in reading message from Web Sockets", "Error: ", err)
			impl.HandleDisconnection(conn, userEmail)
			break
		}
		ctx := context.Background()
		err = impl.client.Publish(ctx, fmt.Sprintf("%s_mobile", userEmail), string(message)).Err()
		if err != nil {
			impl.logger.Errorw("Error in publishing message to Redis", "Error: ", err)
			continue
		}
	}
}

func (impl *LinksImpl) WriteMessages(conn *websocket.Conn, userEmail string) {
	// Write message to clients from Redis
	ctx := context.Background()
	pubSub := impl.client.Subscribe(ctx, fmt.Sprintf("%s_web", userEmail))

	for {
		msg, err := pubSub.ReceiveMessage(ctx)
		if err != nil {
			impl.logger.Errorw("Error in receiving message from pubSub", "Error: ", err)
			_ = conn.WriteMessage(websocket.TextMessage, []byte("Error in receiving message from Database. Try again."))
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
		if err != nil {
			impl.logger.Errorw("Error in writing message to Web Sockets", "Error: ", err)
			impl.HandleDisconnection(conn, userEmail)
			_ = pubSub.Close()
			break
		}

		err = impl.client.Del(ctx, msg.Channel).Err()
		if err != nil {
			impl.logger.Errorw("Error deleting message from Redis", "Error: ", err)
		}
	}

}
