package util

import (
	"go.uber.org/zap"
	"log"
	"runtime/debug"
)

type Async struct {
	logger *zap.SugaredLogger
}

func NewAsync(logger *zap.SugaredLogger) *Async {
	return &Async{
		logger: logger,
	}
}

func (impl *Async) Run(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if impl.logger == nil {
					log.Println("go-routine recovered from panic", "err:", r, "stack:", string(debug.Stack()))
				} else {
					impl.logger.Errorw("go-routine recovered from panic", "err", r, "stack", string(debug.Stack()))
				}
			}
		}()
		if fn != nil {
			fn()
		}
	}()
}
