package services

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/iraunit/get-link-backend/pkg/fileManager"
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
	VerifyEmail(message string, sender string)
	ParseMessageAndBroadcast(message string, sender string) error
}

type WhatsappServiceImpl struct {
	logger       *zap.SugaredLogger
	cfg          bean.WhatsAppConfig
	restClient   util.RestClient
	mailService  MailService
	tokenService TokenService
	repository   repository.Repository
	linkService  LinkService
	fileManager  fileManager.FileManager
}

func NewWhatsappServiceImpl(logger *zap.SugaredLogger, restClient util.RestClient, mailService MailService, tokenService TokenService, repository repository.Repository, linkService LinkService, fileManager fileManager.FileManager) *WhatsappServiceImpl {
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
		repository:   repository,
		linkService:  linkService,
		fileManager:  fileManager,
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
	_, err := impl.restClient.SendWhatsappMessage(fmt.Sprintf(util.WhatsappCloudApiSendMessage, impl.cfg.PhoneID), message)
	return err
}

func (impl *WhatsappServiceImpl) ReceiveMessage(message *bean.WhatsAppBusinessMessageData) error {
	if message.Type == "text" {
		pattern := `^set\s+email\s+\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b$`
		regex := regexp.MustCompile(pattern)
		if regex.MatchString(strings.ToLower(message.Text.Body)) {
			impl.VerifyEmail(message.Text.Body, message.From)
			_ = impl.SendMessage(message.From, "Thanks for using Get-Link. We have sent you an email for verification. Please verify your email. \n\nYou can share your feedback or report an issue on codingkaro.in.\n\nRegards\nRaunit Verma\nShypt Solution")
		} else {
			return impl.ParseMessageAndBroadcast(message.Text.Body, message.From)
		}
	} else if message.Type == "image" {
		id := message.Image.ID
		data, err := impl.getMediaData(id)
		if err != nil {
			return err
		}
		err = impl.downloadMedia(data.Url, data.MimeType, message.From, "data.FileName")
		if err != nil {
			impl.logger.Errorw("Error in downloading media", "Error", err)
			return err
		}
	} else if message.Type == "document" {

	} else if message.Type == "video" {

	}
	return nil
}

func (impl *WhatsappServiceImpl) getMediaData(id string) (*bean.WhatsappMedia, error) {

	mediaData, err := impl.restClient.GetMediaDataFromId(fmt.Sprintf(util.WhatsappCloudApiGetMediaDataUrl, id))

	if err != nil {
		impl.logger.Errorw("Error in getting whatsapp media data", "Error", err)
		return nil, err
	}

	return mediaData, nil
}

func (impl *WhatsappServiceImpl) downloadMedia(url, mimeType, sender, fileName string) error {

	fileExtension, err := util.GetFileExtension(mimeType)
	if err != nil {
		impl.logger.Errorw("Error in getting file extension", "Error", err)
		return err
	}

	allEmails, err := impl.GetUsersFromWhatsappNumber(sender)
	if err != nil {
		impl.logger.Errorw("Error in getting user from whatsapp number", "Error", err)
		return err
	}

	for _, email := range allEmails {
		decryptedEmail, err := util.DecryptData(sender, email.Email, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
			return err
		}
		folderPath := impl.fileManager.GetPathToSaveFileFromWhatsapp(util.EncodeString(email.Email))
		impl.fileManager.DeleteFileFromPathOlderThan24Hours(folderPath)
		folderSize, err := impl.fileManager.GetSizeOfADirectory(folderPath)
		if err != nil {
			impl.logger.Errorw("Error in getting folder size", "Error", err)
			return err
		}
		maxLimit := util.FreeWhatsappFileLimitSizeMB
		if impl.repository.IsUserPremiumUser(decryptedEmail) {
			maxLimit = util.PremiumWhatsappFileLimitSizeMB
		}

		if folderSize > int64(maxLimit) {
			impl.fileManager.DeleteAllFileFromPath(folderPath)
		}

		fileName += "." + fileExtension + ".bin"

		impl.restClient.DownloadMediaFromUrl(url, impl.cfg.AuthToken, fmt.Sprintf("%s/%s", folderPath, fileName), decryptedEmail)
	}

	return nil
}

func (impl *WhatsappServiceImpl) VerifyEmail(message string, number string) {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emails := re.FindAllString(message, -1)
	claims := bean.WhatsappVerificationClaims{Email: emails[0], WhatAppNumber: number}
	token, err := impl.tokenService.EmailVerificationToken(&claims)
	if err != nil {
		impl.logger.Errorw("Error in generating token", "Error", err)
		return
	}
	emailBody := fmt.Sprintf("Please click on the below link to verify your email. \n\n %s/%s?token=%s", impl.cfg.Baseurl, util.VerifyWhatsappEmail, url.QueryEscape(token))
	err = impl.mailService.SendMail(emails[0], "Get-Link - Email Verification", emailBody)
	if err != nil {
		impl.logger.Errorw("Error in sending mail", "Error", err)
		return
	}
	impl.logger.Infow("Verification Mail Sent", "Email", emails[0], "Number", number)

}

func (impl *WhatsappServiceImpl) ParseMessageAndBroadcast(message string, sender string) error {
	allEmails, err := impl.GetUsersFromWhatsappNumber(sender)
	if err != nil {
		impl.logger.Errorw("Error in getting user from whatsapp number", "Error: ", err)
		return err
	}
	for _, email := range allEmails {
		decryptedEmail, err := util.DecryptData(sender, email.Email, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
			return err
		}
		impl.linkService.AddLink(decryptedEmail, &bean.GetLink{Receiver: decryptedEmail, Sender: decryptedEmail, Message: message, UUID: "whatsapp"})
	}
	return nil
}

func (impl *WhatsappServiceImpl) GetUsersFromWhatsappNumber(sender string) ([]bean.WhatsappEmail, error) {
	encryptedSender, err := util.EncryptData(sender, sender, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return nil, err
	}
	allEmails, err := impl.repository.GetEmailsFromNumber(encryptedSender)
	if err != nil {
		impl.logger.Errorw("Error in getting emails from number", "Error: ", err)
		return nil, err
	}
	if len(allEmails) == 0 {
		return nil, pg.ErrNoRows
	}
	return allEmails, nil
}
