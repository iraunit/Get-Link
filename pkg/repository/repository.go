package repository

import (
	"context"
	"encoding/json"
	"github.com/go-pg/pg"
	"github.com/iraunit/get-link-backend/pkg/cryptography"
	"github.com/iraunit/get-link-backend/util/bean"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
	"time"
)

type User struct {
	Email             string    `sql:"email"`
	Premium           bool      `sql:"premium"`
	PremiumExpiration time.Time `sql:"premium_expiration"`
	WhatsappNumber    string    `sql:"whatsapp_number"`
}

type Repository interface {
	AddLink(getLink *bean.GetLink, decryptedData *bean.GetLink, receiverMail string)
	DeleteLink(data *bean.GetLink) error
	GetAllLink(dst string, uuid string) *[]bean.GetLink
	InsertUpdateWhatsappNumber(claims *bean.WhatsappEmail) error
	GetEmailsFromWhatsappNumber(number string) ([]bean.WhatsappEmail, error)
	GetEmailsFromTelegramSender(sender string) ([]bean.TelegramEmail, error)
	IsUserPremiumUser(userEmail string) bool
	InsertUpdateTelegramNumber(email, chatId, senderId string) error
	GetEmailsFromEmail(sender string) ([]bean.TelegramEmail, error)
	GetWhatsappNumberFromEmail(email string) (string, error)
}

type Impl struct {
	db     *pg.DB
	lock   *sync.Mutex
	logger *zap.SugaredLogger
	client *redis.Client
}

func NewRepositoryImpl(db *pg.DB, logger *zap.SugaredLogger, client *redis.Client) *Impl {
	return &Impl{
		db:     db,
		lock:   &sync.Mutex{},
		logger: logger,
		client: client,
	}
}

func (impl *Impl) AddLink(getLink *bean.GetLink, decryptedData *bean.GetLink, receiverMail string) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	result, err := impl.db.Model(getLink).Insert()
	if result.RowsAffected() > 0 {
		pubSubMessage := bean.PubSubMessage{Message: decryptedData.Message, UUID: decryptedData.UUID, ID: getLink.ID, Sender: decryptedData.Sender}
		pubSubMessageJson, err := json.Marshal(pubSubMessage)
		encryptedJson, err := cryptography.EncryptData(receiverMail, string(pubSubMessageJson), impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting json", "Error: ", err)
			return
		} else {
			_ = impl.client.Publish(context.Background(), getLink.Receiver, encryptedJson).Err()
		}
	}
	if err != nil {
		impl.logger.Errorw("Error in adding link", "Error: ", err)
	}
}

func (impl *Impl) DeleteLink(data *bean.GetLink) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	_, err := impl.db.Model(data).WherePK().Delete()
	if err != nil {
		impl.logger.Errorw("Error in deleting link", "Error: ", err)
		return err
	}
	return nil
}

func (impl *Impl) GetAllLink(receiver string, uuid string) *[]bean.GetLink {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	var result []bean.GetLink
	impl.logger.Infow("Info", "Receiver", receiver, "UUID", uuid)
	err := impl.db.Model(&result).
		Column("id", "sender", "message", "uuid").
		Where("receiver=?", receiver).
		Where("uuid != ?", uuid).
		Select()
	if err != nil {
		impl.logger.Errorw("Error in getting link", "Error: ", err)
		return nil
	}
	return &result
}

func (impl *Impl) InsertUpdateWhatsappNumber(claims *bean.WhatsappEmail) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	var prevRecord []bean.WhatsappEmail
	err := impl.db.Model(&prevRecord).Where("email = ?", claims.Email).Select()
	if err != nil {
		impl.logger.Errorw("Error in verifying link", "Error: ", err)
		return err
	}
	if prevRecord == nil {
		_, err = impl.db.Model(claims).Insert()
		if err != nil {
			impl.logger.Errorw("Error in verifying link", "Error: ", err)
		}
		return err
	} else {
		_, err = impl.db.Model(claims).Where("email=?", claims.Email).Update()
		if err != nil {
			impl.logger.Errorw("Error in verifying link", "Error: ", err)
		}
		return err
	}
}

func (impl *Impl) GetEmailsFromWhatsappNumber(number string) ([]bean.WhatsappEmail, error) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	var result []bean.WhatsappEmail
	err := impl.db.Model(&result).Column("email").Where("whatsapp_number = ?", number).Select()
	if err != nil {
		impl.logger.Errorw("Error in getting emails from number", "Error: ", err)
		return nil, err
	}
	return result, nil
}

func (impl *Impl) IsUserPremiumUser(userEmail string) bool {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	var user User
	err := impl.db.Model(&user).Where("email = ?", userEmail).Select()
	if err != nil || user.Email == "" {
		impl.logger.Errorw("Error in getting user", "Error: ", err, "user", user)
		return false
	}

	if user.Premium && user.PremiumExpiration.After(time.Now()) {
		return true
	}
	return false
}

func (impl *Impl) InsertUpdateTelegramNumber(email, chatId, senderId string) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	_, err := impl.db.Model(&bean.TelegramEmail{}).Where("email = ?", email).Delete()
	if err != nil {
		return err
	}
	_, err = impl.db.Model(&bean.TelegramEmail{Email: email, ChatId: chatId, SenderId: senderId}).Insert()
	return err
}

func (impl *Impl) GetEmailsFromTelegramSender(sender string) ([]bean.TelegramEmail, error) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	var result []bean.TelegramEmail
	err := impl.db.Model(&result).Column("email").Where("sender_id = ?", sender).Select()
	if err != nil {
		impl.logger.Errorw("Error in getting emails from number", "Error: ", err)
		return nil, err
	}
	return result, nil
}

func (impl *Impl) GetEmailsFromEmail(sender string) ([]bean.TelegramEmail, error) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	var result []bean.TelegramEmail
	err := impl.db.Model(&result).Where("email = ?", sender).Select()
	if err != nil {
		impl.logger.Errorw("Error in getting emails from email", "Error: ", err)
		return nil, err
	}
	return result, nil
}

func (impl *Impl) GetWhatsappNumberFromEmail(email string) (string, error) {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	var user User
	err := impl.db.Model(&user).Column("whatsapp_number").Where("email = ?", email).Select()
	if err != nil {
		impl.logger.Errorw("Error in getting whatsapp number from email", "Error: ", err)
		return "", err
	}
	return user.WhatsappNumber, nil
}
