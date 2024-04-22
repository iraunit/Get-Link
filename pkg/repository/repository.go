package repository

import (
	"fmt"
	"github.com/go-pg/pg"
	"github.com/iraunit/get-link-backend/util"
	"go.uber.org/zap"
	"sync"
)

type Repository interface {
	AddLink(getLink *util.GetLink)
	DeleteLink(data *util.GetLink) error
	GetAllLink(dst string, uuid string) *[]util.GetLink
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

func (impl *Impl) AddLink(getLink *util.GetLink) {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	_, err := impl.db.Model(getLink).Insert()

	if err != nil {
		impl.logger.Errorw("Error in adding link", "Error: ", err)
	}
}

func (impl *Impl) DeleteLink(data *util.GetLink) error {
	impl.lock.Lock()
	defer impl.lock.Unlock()
	_, err := impl.db.Model(data).WherePK().Delete()
	if err != nil {
		impl.logger.Errorw("Error in deleting link", "Error: ", err)
		return err
	}
	return nil
}

func (impl *Impl) GetAllLink(dst string, uuid string) *[]util.GetLink {
	impl.lock.Lock()
	defer impl.lock.Unlock()

	var result []util.GetLink
	impl.logger.Infow("Info", "Destination", dst, "UUID", uuid)
	err := impl.db.Model(&result).Where("destination=?", dst).Where("uuid!=?", uuid).Select()
	if err != nil {
		impl.logger.Errorw("Error in getting link", "Error: ", err)
		return nil
	}
	fmt.Println(result)
	return &result
}
