package services

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/smtp"
)

type MailService interface {
	SendMail(receiver string, subject string, body string) error
}

type MailServiceImpl struct {
	logger *zap.SugaredLogger
	cfg    bean.MailConfig
}

func NewMailServiceImpl(logger *zap.SugaredLogger) *MailServiceImpl {
	cfg := bean.MailConfig{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}
	return &MailServiceImpl{
		logger: logger,
		cfg:    cfg,
	}
}

func (impl *MailServiceImpl) SendMail(receiver string, subject string, body string) error {
	auth := smtp.PlainAuth("", impl.cfg.From, impl.cfg.Password, impl.cfg.Host)
	msg := "From: " + impl.cfg.From + "\n" +
		"To: " + receiver + "\n" +
		"Subject: " + subject + "\n\n" +
		body
	err := smtp.SendMail(fmt.Sprintf("%s:%s", impl.cfg.Host, impl.cfg.Port), auth, impl.cfg.From, []string{receiver}, []byte(msg))
	return err
}
