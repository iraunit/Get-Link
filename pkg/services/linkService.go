package services

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/util"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
)

type LinkService interface {
	ReadMessages(conn *websocket.Conn, userEmail string, uuid string, device string)
	WriteMessages(conn *websocket.Conn, userEmail string, uuid string, device string)
	HandleDisconnection(conn *websocket.Conn, userEmail string)
	HandleConnection(conn *websocket.Conn, userEmail string)
	AddLink(userEmail string, data *util.GetLink)
	GetAllLink(userEmail string, dst string, uuid string) *[]util.GetLink
	DeleteLink(userEmail string, data *util.GetLink) error
}

type LinkServiceImpl struct {
	logger     *zap.SugaredLogger
	client     *redis.Client
	lock       *sync.Mutex
	Users      *map[string]util.User
	Repository repository.Repository
}

func NewLinkServiceImpl(client *redis.Client, logger *zap.SugaredLogger, users *map[string]util.User, repository repository.Repository) *LinkServiceImpl {
	return &LinkServiceImpl{
		logger:     logger,
		client:     client,
		lock:       &sync.Mutex{},
		Users:      users,
		Repository: repository,
	}
}

func (impl *LinkServiceImpl) ReadMessages(conn *websocket.Conn, userEmail string, uuid string, device string) {
	//Read message from Client and push to Redis
	destination := device
	if destination == "web" {
		destination = "mobile"
	} else {
		destination = "web"
	}
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			impl.logger.Errorw("Error in reading message from Web Sockets", "Error: ", err)
			impl.HandleDisconnection(conn, userEmail)
			break
		}
		ctx := context.Background()
		encryptedMsg, err := util.EncryptData(userEmail, string(message), impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encryption", "Error: ", err)
			continue
		}
		data := util.GetLink{
			Destination: fmt.Sprintf("%s_%s", encryptedEmail, destination),
			Message:     encryptedMsg,
			UUID:        uuid,
		}
		impl.Repository.AddLink(&data)
		err = impl.client.Publish(ctx, fmt.Sprintf("%s_%s", encryptedEmail, destination), encryptedMsg).Err()
		if err != nil {
			impl.logger.Errorw("Error in publishing message to Redis", "Error: ", err)
			continue
		}
	}
}

func (impl *LinkServiceImpl) WriteMessages(conn *websocket.Conn, userEmail string, uuid string, device string) {
	// Write message to clients from Redis
	destination := device
	
	ctx := context.Background()
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return
	}
	pubSub := impl.client.Subscribe(ctx, fmt.Sprintf("%s_%s", encryptedEmail, destination))

	for {
		msg, err := pubSub.ReceiveMessage(ctx)
		if err != nil {
			impl.logger.Errorw("Error in receiving message from pubSub", "Error: ", err)
			_ = conn.WriteMessage(websocket.TextMessage, []byte("Error in receiving message from Database. Try again."))
			continue
		}
		decryptedMsg, err := util.DecryptData(userEmail, msg.Payload, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decryption", "Error: ", err)
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(decryptedMsg))
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

func (impl *LinkServiceImpl) AddLink(userEmail string, data *util.GetLink) {
	encryptedMail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return
	}
	encryptedData, err := util.EncryptData(userEmail, data.Message, impl.logger)

	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return
	}

	data.Message = encryptedData
	data.Destination = fmt.Sprintf("%s_%s", encryptedMail, data.Destination)

	impl.Repository.AddLink(data)
}

func (impl *LinkServiceImpl) HandleConnection(conn *websocket.Conn, userEmail string) {
	impl.lock.Lock()
	user, ok := (*impl.Users)[userEmail]
	if !ok {
		user = util.User{
			Lock:        &sync.Mutex{},
			Connections: make([]*websocket.Conn, 0),
		}
		(*impl.Users)[userEmail] = user
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

func (impl *LinkServiceImpl) HandleDisconnection(conn *websocket.Conn, userEmail string) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	user, ok := (*impl.Users)[userEmail]

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
		delete(*impl.Users, userEmail)
	}
}

func (impl *LinkServiceImpl) GetAllLink(userEmail string, dst string, uuid string) *[]util.GetLink {
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return nil
	}
	return impl.Repository.GetAllLink(fmt.Sprintf("%s_%s", encryptedEmail, dst), uuid)
}

func (impl *LinkServiceImpl) DeleteLink(userEmail string, data *util.GetLink) error {
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return err
	}
	data.Destination = fmt.Sprintf("%s_%s", encryptedEmail, "web")
	data.Message = ""
	err = impl.Repository.DeleteLink(data)
	if err != nil {
		return err
	}
	data.Destination = fmt.Sprintf("%s_%s", encryptedEmail, "mobile")
	err = impl.Repository.DeleteLink(data)
	if err != nil {
		return err
	}
	return nil
}