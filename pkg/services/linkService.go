package services

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
)

type LinkService interface {
	ReadMessages(conn *websocket.Conn, userEmail string, uuid string)
	WriteMessages(conn *websocket.Conn, userEmail string)
	HandleDisconnection(conn *websocket.Conn, userEmail string)
	HandleConnection(conn *websocket.Conn, userEmail string)
	AddLink(userEmail string, data *bean.GetLink)
	GetAllLink(userEmail string, uuid string) *[]bean.GetLink
	DeleteLink(userEmail string, data *bean.GetLink) error
	VerifyWhatsapp(userEmail string, claims *bean.WhatsappEmail) error
}

type LinkServiceImpl struct {
	logger     *zap.SugaredLogger
	client     *redis.Client
	lock       *sync.Mutex
	Users      *map[string]bean.User
	Repository repository.Repository
}

func NewLinkServiceImpl(client *redis.Client, logger *zap.SugaredLogger, users *map[string]bean.User, repository repository.Repository) *LinkServiceImpl {

	return &LinkServiceImpl{
		logger:     logger,
		client:     client,
		lock:       &sync.Mutex{},
		Users:      users,
		Repository: repository,
	}
}

func (impl *LinkServiceImpl) ReadMessages(conn *websocket.Conn, userEmail string, uuid string) {
	//Read message from Client and push to Redis
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			var closeErr *websocket.CloseError
			if errors.As(err, &closeErr) {
				if closeErr.Code == websocket.CloseGoingAway {
					impl.HandleDisconnection(conn, userEmail)
					return
				}
			}
			impl.logger.Errorw("Error in reading message from Web Sockets", "Error: ", err)
			impl.HandleDisconnection(conn, userEmail)
			return
		}
		ctx := context.Background()
		encryptedMsg, err := util.EncryptData(userEmail, string(message), impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encryption", "Error: ", err)
			continue
		}

		data := bean.GetLink{
			Sender:   encryptedEmail,
			Receiver: encryptedEmail,
			Message:  encryptedMsg,
			UUID:     uuid,
		}
		impl.Repository.AddLink(&data, &bean.GetLink{Sender: userEmail, Receiver: userEmail, Message: string(message), UUID: uuid}, userEmail)
		err = impl.client.Publish(ctx, encryptedEmail, encryptedMsg).Err()
		if err != nil {
			impl.logger.Errorw("Error in publishing message to Redis", "Error: ", err)
			continue
		}
	}
}

func (impl *LinkServiceImpl) WriteMessages(conn *websocket.Conn, userEmail string) {
	// Write message to clients from Redis

	ctx := context.Background()
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return
	}
	pubSub := impl.client.Subscribe(ctx, encryptedEmail)

	for {
		msg, err := pubSub.ReceiveMessage(ctx)
		if err != nil {
			impl.logger.Errorw("Error in receiving message from pubSub", "Error: ", err)
			_ = conn.WriteMessage(websocket.TextMessage, []byte("Error in receiving message from Database. Try again."))
			impl.HandleDisconnection(conn, userEmail)
			return
		}
		decryptedMsg, err := util.DecryptData(userEmail, msg.Payload, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decryption", "Error: ", err)
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(decryptedMsg))
		if err != nil {
			impl.logger.Errorw("Error in writing message to Web Sockets", "Error: ", err)
			_ = pubSub.Close()
			impl.HandleDisconnection(conn, userEmail)
			return
		}

		err = impl.client.Del(ctx, msg.Channel).Err()
		if err != nil {
			impl.logger.Errorw("Error deleting message from Redis", "Error: ", err)
		}
	}

}

func (impl *LinkServiceImpl) AddLink(userEmail string, data *bean.GetLink) {
	receiverMail := data.Receiver
	decryptedData := bean.GetLink{ID: data.ID, Sender: data.Sender, Receiver: data.Receiver, Message: data.Message, UUID: data.UUID}
	receiverMailEncrypted, err := util.EncryptData(data.Receiver, data.Receiver, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return
	}
	senderMailEncrypted, err := util.EncryptData(data.Receiver, data.Sender, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return
	}

	encryptedData, err := util.EncryptData(userEmail, data.Message, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return
	}

	data.Sender = senderMailEncrypted
	data.Receiver = receiverMailEncrypted
	data.Message = encryptedData

	impl.Repository.AddLink(data, &decryptedData, receiverMail)
}

func (impl *LinkServiceImpl) HandleConnection(conn *websocket.Conn, userEmail string) {
	impl.lock.Lock()
	user, ok := (*impl.Users)[userEmail]
	if !ok {
		user = bean.User{
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

func (impl *LinkServiceImpl) GetAllLink(userEmail string, uuid string) *[]bean.GetLink {
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return nil
	}

	allLinks := impl.Repository.GetAllLink(encryptedEmail, uuid)
	for i := 0; i < len(*allLinks); i++ {
		(*allLinks)[i].Sender, err = util.DecryptData(userEmail, (*allLinks)[i].Sender, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
		}
		(*allLinks)[i].Receiver, err = util.DecryptData(userEmail, (*allLinks)[i].Receiver, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
		}
		(*allLinks)[i].Message, err = util.DecryptData(userEmail, (*allLinks)[i].Message, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
		}
	}
	return allLinks
}

func (impl *LinkServiceImpl) DeleteLink(userEmail string, data *bean.GetLink) error {
	encryptedEmail, err := util.EncryptData(userEmail, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return err
	}
	data.Receiver = encryptedEmail
	return impl.Repository.DeleteLink(data)
}
func (impl *LinkServiceImpl) VerifyWhatsapp(userEmail string, claims *bean.WhatsappEmail) error {
	sender := claims.WhatAppNumber
	encryptedEmail, err := util.EncryptData(sender, userEmail, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return err
	}
	claims.Email = encryptedEmail
	if claims.WhatAppNumber != "" {
		encryptedWhatsapp, err := util.EncryptData(sender, claims.WhatAppNumber, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting data", "Error: ", err)
			return err
		}
		claims.WhatAppNumber = encryptedWhatsapp
	}
	return impl.Repository.InsertUpdateWhatsappNumber(claims)
}
