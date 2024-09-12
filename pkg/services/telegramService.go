package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/iraunit/get-link-backend/pkg/cryptography"
	"github.com/iraunit/get-link-backend/pkg/fileManager"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/pkg/restCalls"
	tokenService2 "github.com/iraunit/get-link-backend/pkg/services/tokenService"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"regexp"

	"strings"
)

type TelegramFileResponse struct {
	Ok     bool `json:"ok"`
	Result File `json:"result"`
}

type File struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     int    `json:"file_size,omitempty"`
	FilePath     string `json:"file_path,omitempty"`
	FileName     string `json:"file_name,omitempty"`
}

type TelegramService interface {
	ReceiveTelegramMessage(ctx context.Context, b *bot.Bot, update *models.Update)
	SendTelegramMessage(chatID int64, message string)
	VerifyTelegram(email string, claims *bean.TelegramVerificationClaims) error
	GetUsersFromTelegramNumber(sender string) ([]bean.TelegramEmail, error)
	GetUsersFromEmail(email string) ([]bean.TelegramEmail, error)
}

type TelegramImpl struct {
	logger       *zap.SugaredLogger
	async        *util.Async
	bot          *bot.Bot
	ctx          context.Context
	mailService  MailService
	tokenService tokenService2.TokenService
	cfg          *bean.TelegramCfg
	repository   repository.Repository
	linkService  LinkService
	fileManager  fileManager.FileManager
	restClient   restCalls.RestClient
}

func NewTelegramService(logger *zap.SugaredLogger, async *util.Async, mailService MailService, tokenService tokenService2.TokenService, repository repository.Repository, linkService LinkService, fileManager fileManager.FileManager, restClient restCalls.RestClient) *TelegramImpl {
	ctx := context.Background()
	cfg := &bean.TelegramCfg{}
	if err := env.Parse(cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}
	impl := &TelegramImpl{
		logger:       logger,
		async:        async,
		ctx:          ctx,
		cfg:          cfg,
		mailService:  mailService,
		tokenService: tokenService,
		repository:   repository,
		fileManager:  fileManager,
		linkService:  linkService,
		restClient:   restClient,
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(impl.ReceiveTelegramMessage),
	}

	b, err := bot.New(cfg.TelegramToken, opts...)
	if err != nil {
		logger.Fatalw("error in creating telegram bot", "error", err)
	}

	impl.bot = b

	async.Run(func() {
		b.Start(ctx)
	})
	return impl
}

func (impl *TelegramImpl) ReceiveTelegramMessage(ctx context.Context, b *bot.Bot, update *models.Update) {
	impl.bot = b
	if strings.Contains(strings.ToLower(update.Message.Text), "set email") {
		// send email here
		re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
		emails := re.FindAllString(update.Message.Text, -1)
		if len(emails) == 0 {
			impl.SendTelegramMessage(update.Message.Chat.ID, "Please enter valid email")
			return
		}
		claims := bean.TelegramVerificationClaims{Email: emails[0], ChatId: update.Message.Chat.ID, SenderId: update.Message.From.ID}
		token, err := impl.tokenService.TelegramEmailVerificationToken(&claims)
		if err != nil {
			impl.logger.Errorw("Error in generating token", "Error", err)
			impl.SendTelegramMessage(update.Message.Chat.ID, "error in generating token. Please contact support at https://x.com/iraunit")
			return
		}
		emailBody := fmt.Sprintf("Please click on the below link to verify your email. \n\n %s%s?token=%s", impl.cfg.BaseUrl, util.VerifyTelegramEmail, url.QueryEscape(token))
		err = impl.mailService.SendMail(emails[0], "Get-Link - Email Verification", emailBody)
		if err != nil {
			impl.logger.Errorw("Error in sending mail", "Error", err)
			impl.SendTelegramMessage(update.Message.Chat.ID, "Error in sending verification mail. Please contact support at https://x.com/iraunit")
			return
		}
		impl.SendTelegramMessage(update.Message.Chat.ID, "Please check your email for verification link. \n\nYou can share your feedback or report an issue on codingkaro.in.\n\nRegards\nRaunit Verma\nShypt Solution")

	} else if update.Message.Text != "" {
		allEmails, err := impl.GetUsersFromTelegramNumber(strconv.FormatInt(update.Message.From.ID, 10))
		if err != nil {
			impl.logger.Errorw("Error in getting users from telegram number", "Error", err)
			impl.SendTelegramMessage(update.Message.Chat.ID, "Have you set your email here. Please send 'set email youremail@gmail.com' and then verify by clicking on the link received on your email.")
			return
		}
		for _, email := range allEmails {
			decryptedEmail, err := cryptography.DecryptData(strconv.FormatInt(update.Message.From.ID, 10), email.Email, impl.logger)
			if err != nil {
				impl.logger.Errorw("Error in decrypting data", "Error", err)
				impl.SendTelegramMessage(update.Message.Chat.ID, "Error in decrypting data")
				return
			}
			impl.linkService.AddLink(decryptedEmail, &bean.GetLink{Receiver: decryptedEmail, Sender: decryptedEmail, Message: update.Message.Text, UUID: "telegram"})
		}
		impl.SendTelegramMessage(update.Message.Chat.ID, update.Message.Text)
	} else {
		if update.Message.Photo != nil && len(update.Message.Photo) > 0 {
			flag := true
			for _, photo := range update.Message.Photo {
				filePath, err := impl.getFilePath(photo.FileID)
				if err != nil {
					impl.logger.Errorw("Error in getting file path", "Error", err)
					impl.SendTelegramMessage(update.Message.Chat.ID, "Error in getting file path, cannot send image to get link devices.")
					return
				}
				fileArray := strings.Split(filePath, "/")
				fileName := fileArray[len(fileArray)-1]
				err = impl.downloadMedia(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", impl.cfg.TelegramToken, filePath), strconv.FormatInt(update.Message.From.ID, 10), fileName)
				if err != nil {
					impl.logger.Errorw("Error in downloading media", "Error", err)
					impl.SendTelegramMessage(update.Message.Chat.ID, "error in downloading media, cannot send image to get link devices.")
					flag = false
				}
			}
			if flag {
				impl.SendTelegramMessage(update.Message.Chat.ID, "Image uploaded successfully. \n\nYou can share your feedback or report an issue on codingkaro.in.\n\nRegards\nRaunit Verma\nShypt Solution")
			}
		} else {
			switch {
			case update.Message.Document != nil:
				impl.handleMedia("Document", update.Message.Document.FileID, update.Message.Document.FileName, update.Message.Chat.ID, update.Message.From.ID)
			case update.Message.Audio != nil:
				impl.handleMedia("Audio", update.Message.Audio.FileID, update.Message.Audio.FileName, update.Message.Chat.ID, update.Message.From.ID)
			case update.Message.Video != nil:
				impl.handleMedia("Video", update.Message.Video.FileID, update.Message.Video.FileName, update.Message.Chat.ID, update.Message.From.ID)
			}
		}
	}
}

func (impl *TelegramImpl) handleMedia(uploadType string, fileID, fileName string, chatID, userID int64) {
	filePath, err := impl.getFilePath(fileID)
	if err != nil {
		impl.logger.Errorw(fmt.Sprintf("Error in getting file path for %s", uploadType), "Error", err)
		impl.SendTelegramMessage(chatID, fmt.Sprintf("Error in getting file path, cannot send %s to get link devices.", uploadType))
		return
	}
	if fileName == "" {
		fileNameArray := strings.Split(filePath, "/")
		fileName = fileNameArray[len(fileNameArray)-1]
	}
	err = impl.downloadMedia(fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", impl.cfg.TelegramToken, filePath), strconv.FormatInt(userID, 10), fileName)
	if err != nil {
		impl.logger.Errorw(fmt.Sprintf("Error in downloading %s", uploadType), "Error", err)
		impl.SendTelegramMessage(chatID, fmt.Sprintf("Error in downloading %s, cannot send %s to get link devices.", uploadType, uploadType))
	} else {
		impl.SendTelegramMessage(chatID, fmt.Sprintf("%s uploaded successfully.\n\nYou can share your feedback or report an issue on codingkaro.in.\n\nRegards\nRaunit Verma\nShypt Solution", uploadType))
	}
}

func (impl *TelegramImpl) SendTelegramMessage(chatID int64, message string) {
	if impl.bot == nil {
		return
	}
	_, err := impl.bot.SendMessage(impl.ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   message,
	})
	if err != nil {
		impl.logger.Errorw("error in sending telegram message", "error", err)
	}
}

func (impl *TelegramImpl) VerifyTelegram(email string, claims *bean.TelegramVerificationClaims) error {
	senderId := strconv.FormatInt(claims.SenderId, 10)
	chatId := strconv.FormatInt(claims.ChatId, 10)
	senderIdCpy := senderId
	chatIdCpy := chatId
	encryptedEmail, err := cryptography.EncryptData(senderId, email, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return err
	}

	if claims.SenderId != 0 {
		senderId, err = cryptography.EncryptData(senderId, senderId, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting data", "Error: ", err)
			return err
		}
	}
	if claims.ChatId != 0 {
		chatId, err = cryptography.EncryptData(senderId, chatId, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting data", "Error: ", err)
			return err
		}
	}
	claims.Email = encryptedEmail
	err = impl.repository.InsertUpdateTelegramNumber(encryptedEmail, chatId, senderId)
	if err != nil {
		impl.logger.Errorw("Error in inserting data", "Error: ", err)
		return err
	}

	encryptedEmail, err = cryptography.EncryptData(email, email, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encrypting data", "Error: ", err)
		return err
	}

	if claims.SenderId != 0 {
		senderId, err = cryptography.EncryptData(email, senderIdCpy, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting data", "Error: ", err)
			return err
		}
	}
	if claims.ChatId != 0 {
		chatId, err = cryptography.EncryptData(email, chatIdCpy, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting data", "Error: ", err)
			return err
		}
	}
	claims.Email = encryptedEmail
	err = impl.repository.InsertUpdateTelegramNumber(encryptedEmail, chatId, senderId)
	if err != nil {
		impl.logger.Errorw("Error in inserting data", "Error: ", err)
		return err
	}

	return nil
}

func (impl *TelegramImpl) GetUsersFromTelegramNumber(sender string) ([]bean.TelegramEmail, error) { // 2121983277
	encryptedSender, err := cryptography.EncryptData(sender, sender, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return nil, err
	}
	allEmails, err := impl.repository.GetEmailsFromTelegramSender(encryptedSender)
	if err != nil {
		impl.logger.Errorw("Error in getting emails from number", "Error: ", err)
		return nil, err
	}
	if len(allEmails) == 0 {
		return nil, pg.ErrNoRows
	}
	return allEmails, nil
}

func (impl *TelegramImpl) GetUsersFromEmail(email string) ([]bean.TelegramEmail, error) {
	encryptedSender, err := cryptography.EncryptData(email, email, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error in encryption", "Error: ", err)
		return nil, err
	}
	allEmails, err := impl.repository.GetEmailsFromEmail(encryptedSender)
	if err != nil {
		impl.logger.Errorw("Error in getting emails from number", "Error: ", err)
		return nil, err
	}
	if len(allEmails) == 0 {
		return nil, pg.ErrNoRows
	}
	return allEmails, nil
}

func (impl *TelegramImpl) getFilePath(fileID string) (string, error) {

	telegramApiUrl := fmt.Sprintf("https://api.telegram.org/bot%s/getFile?file_id=%s", impl.cfg.TelegramToken, fileID)

	resp, err := http.Get(telegramApiUrl)
	if err != nil {
		impl.logger.Errorw("Error in sending request", "Error: ", err)
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			impl.logger.Errorw("Error in closing body", "Error: ", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		impl.logger.Errorw("Error in reading body", "Error: ", err)
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var fileResponse TelegramFileResponse
	err = json.Unmarshal(body, &fileResponse)
	if err != nil {
		impl.logger.Errorw("Error in parsing JSON", "Error: ", err)
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	if !fileResponse.Ok {
		impl.logger.Errorw("Error in getting file path", "Error: ", err)
		return "", fmt.Errorf("failed to get file path: %s", body)
	}
	return fileResponse.Result.FilePath, nil
}

func (impl *TelegramImpl) downloadMedia(url, sender, fileNameWithExtension string) error {

	allEmails, err := impl.GetUsersFromTelegramNumber(sender)
	if err != nil {
		impl.logger.Errorw("Error in getting user from whatsapp number", "Error", err)
		return err
	}

	for _, email := range allEmails {
		decryptedEmail, err := cryptography.DecryptData(sender, email.Email, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
			return err
		}
		folderPath := impl.fileManager.GetPathToSaveFileFromTelegram(util.EncodeString(decryptedEmail))
		impl.fileManager.DeleteFileFromPathOlderThan24Hours(folderPath)
		folderSize, err := impl.fileManager.GetSizeOfADirectory(folderPath)
		if err != nil {
			impl.logger.Errorw("Error in getting folder size", "Error", err)
			return err
		}
		maxLimit := util.FreeTelegramFileLimitSizeMB
		if impl.repository.IsUserPremiumUser(email.Email) {
			maxLimit = util.PremiumTelegramFileLimitSizeMB
		}

		if folderSize > int64(maxLimit) {
			impl.fileManager.DeleteAllFileFromPath(folderPath)
		}

		impl.restClient.DownloadTelegramMediaFromUrl(url, path.Join(folderPath, fileNameWithExtension+".bin"), decryptedEmail)
	}

	return nil
}
