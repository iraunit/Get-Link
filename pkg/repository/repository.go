package repository

import (
	"github.com/go-pg/pg"
	"github.com/iraunit/get-link-backend/util"
	"go.uber.org/zap"
	"sync"
)

type Repository interface {
	AddLink(userMessage *util.UserMessage)
}

type Impl struct {
	db     *pg.DB
	lock   *sync.Mutex
	logger *zap.SugaredLogger
}

func NewRepositoryImpl(db *pg.DB, logger *zap.SugaredLogger) *Impl {
	return &Impl{
		db:     db,
		lock:   &sync.Mutex{},
		logger: logger,
	}
}

func (impl *Impl) AddLink(userMessage *util.UserMessage) {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	_, err := impl.db.Model(userMessage).Insert()

	if err != nil {
		impl.logger.Errorw("Error in adding link", "Error: ", err)
	}
}
