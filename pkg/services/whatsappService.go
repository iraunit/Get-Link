package services

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/iraunit/get-link-backend/pkg/repository"
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
	cryptoConfig bean.EncryptDecryptConfig
	repository   repository.Repository
	linkService  LinkService
}

func NewWhatsappServiceImpl(logger *zap.SugaredLogger, restClient util.RestClient, mailService MailService, tokenService TokenService, repository repository.Repository, linkService LinkService) *WhatsappServiceImpl {
	cfg := bean.WhatsAppConfig{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}

	cryptoConfig := bean.EncryptDecryptConfig{}
	err := env.Parse(&cryptoConfig)
	if err != nil {
		logger.Fatal("Error in parsing config", "Error: ", err)
	}

	return &WhatsappServiceImpl{
		logger:       logger,
		cfg:          cfg,
		restClient:   restClient,
		mailService:  mailService,
		tokenService: tokenService,
		cryptoConfig: cryptoConfig,
		repository:   repository,
		linkService:  linkService,
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
			_ = impl.SendMessage(message.From, "Thanks for using Get-Link. We have sent you an email for verification. Please verify your email. \n\nYou can share your feedback or report an issue on codingkaro.in. \n\nRegards\nRaunit Verma\nShypt Solution")
		} else {
			err := impl.ParseMessageAndBroadcast(message.Text.Body, message.From)
			if err != nil {
				impl.logger.Errorw("Error in parsing message", "Error", err)
				_ = impl.SendMessage(message.From, "Sorry! Something went wrong. Please try again. You can share your feedback or report an issue on codingkaro.in.")
				return err
			} else {
				_ = impl.SendMessage(message.From, "Thanks for using Get-Link. Your request has been processed. You can share your feedback or report an issue on codingkaro.in.")
			}
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
	encryptedSender, err := util.EncryptData(impl.cryptoConfig.EncryptionKey, sender, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return err
	}
	allEmails, err := impl.repository.GetEmailsFromNumber(encryptedSender)
	if err != nil {
		impl.logger.Errorw("Error in getting emails from number", "Error: ", err)
		return err
	}
	for _, email := range allEmails {
		decryptedEmail, err := util.DecryptData(impl.cryptoConfig.EncryptionKey, email.Email, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
			return err
		}
		impl.linkService.AddLink(decryptedEmail, &bean.GetLink{Receiver: decryptedEmail, Sender: decryptedEmail, Message: message, UUID: "whatsapp"})
	}
	return nil
}
