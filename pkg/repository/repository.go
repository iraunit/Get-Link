package repository

import (
	"context"
	"encoding/json"
	"github.com/go-pg/pg"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"sync"
)

type Repository interface {
	AddLink(getLink *bean.GetLink, receiverMail string)
	DeleteLink(data *bean.GetLink) error
	GetAllLink(dst string, uuid string) *[]bean.GetLink
	Verify(claims *bean.UserSocialData) error
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

func (impl *Impl) AddLink(getLink *bean.GetLink, receiverMail string) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	result, err := impl.db.Model(getLink).Insert()
	if result.RowsAffected() > 0 {
		pubSubMessage := bean.PubSubMessage{Message: getLink.Message, UUID: getLink.UUID, ID: getLink.ID, Sender: getLink.Sender}
		pubSubMessageJson, err := json.Marshal(pubSubMessage)
		encryptedJson, err := util.EncryptData(receiverMail, string(pubSubMessageJson), impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting json", "Error: ", err)
			return
		}
		_ = impl.client.Publish(context.Background(), getLink.Receiver, encryptedJson).Err()
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

func (impl *Impl) Verify(claims *bean.UserSocialData) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	_, err := impl.db.Model(claims).WherePK().Update()
	if err != nil {
		impl.logger.Errorw("Error in verifying link", "Error: ", err)
		return err
	}
	return nil
}
