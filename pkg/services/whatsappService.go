package services

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/url"
	"regexp"
	"strings"
)

type WhatsappService interface {
	SendMessage(number string, body string) error
	ReceiveMessage(message *bean.WhatsAppBusinessMessageData) error
	GetMediaLink(id string) (string, error)
	DownloadMedia(url string) (string, error)
	VerifyEmail(message string, sender string)
	ParseMessageAndBroadcast(message string, sender string) error
}

type WhatsappServiceImpl struct {
	logger       *zap.SugaredLogger
	cfg          bean.WhatsAppConfig
	restClient   util.RestClient
	mailService  MailService
	tokenService TokenService
}

func NewWhatsappServiceImpl(logger *zap.SugaredLogger, restClient util.RestClient, mailService MailService, tokenService TokenService) *WhatsappServiceImpl {
	cfg := bean.WhatsAppConfig{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}
	return &WhatsappServiceImpl{
		logger:       logger,
		cfg:          cfg,
		restClient:   restClient,
		mailService:  mailService,
		tokenService: tokenService,
	}
}

func (impl *WhatsappServiceImpl) SendMessage(number string, body string) error {
	message := bean.WhatsAppBusinessSendTextMessage{
		MessagingProduct: "whatsapp",
		To:               number,
		Type:             "text",
		Text: bean.WhatsAppBusinessTextData{
			Body: body,
		},
	}
	_, err := impl.restClient.SendWhatsappMessage(fmt.Sprintf(util.WHATSAPP_CLOUD_API_SEND_MESSAGE, impl.cfg.PhoneID), message)
	return err
}

func (impl *WhatsappServiceImpl) ReceiveMessage(message *bean.WhatsAppBusinessMessageData) error {
	if message.Type == "text" {
		pattern := `^set\s+email\s+\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b$`
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(strings.ToLower(message.Text.Body)) {
			impl.VerifyEmail(message.Text.Body, message.From)
		} else if regex.MatchString(message.Text.Body) {

		}
	} else if message.Type == "image" {

	} else if message.Type == "document" {

	} else if message.Type == "video" {

	}
	return nil
}

func (impl *WhatsappServiceImpl) GetMediaLink(id string) (string, error) {

	return "", nil
}

func (impl *WhatsappServiceImpl) DownloadMedia(url string) (string, error) {

	return "", nil
}

func (impl *WhatsappServiceImpl) VerifyEmail(message string, number string) {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := re.FindAllString(message, -1)
	claims := bean.EmailVerificationClaims{Email: emails[0], WhatAppNumber: number}
	token, err := impl.tokenService.EmailVerificationToken(&claims)
	if err != nil {
		impl.logger.Errorw("Error in generating token", "Error", err)
		return
	}
	emailBody := fmt.Sprintf("Please click on the below link to verify your email. \n\n %s/verifyEmail?token=%s", impl.cfg.Baseurl, url.QueryEscape(token))
	err = impl.mailService.SendMail(emails[0], "Get-Link - Email Verification", emailBody)
	if err != nil {
		impl.logger.Errorw("Error in sending mail", "Error", err)
		return
	}
	impl.logger.Infow("Verification Mail Sent", "Email", emails[0], "Number", number)

}

func (impl *WhatsappServiceImpl) ParseMessageAndBroadcast(message string, sender string) error {
	return nil
}
