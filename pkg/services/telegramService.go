package services

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/iraunit/get-link-backend/pkg/cryptography"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/url"
	"strconv"

	"regexp"

	"strings"
)

type TelegramService interface {
	ReceiveTelegramMessage(ctx context.Context, b *bot.Bot, update *models.Update)
	SendTelegramMessage(chatID int64, message string)
	VerifyTelegram(email string, claims *bean.TelegramVerificationClaims) error
}

type TelegramImpl struct {
	logger       *zap.SugaredLogger
	async        *util.Async
	bot          *bot.Bot
	ctx          context.Context
	mailService  MailService
	tokenService TokenService
	cfg          *bean.TelegramCfg
	repository   repository.Repository
}

func NewTelegramService(logger *zap.SugaredLogger, async *util.Async, mailService MailService, tokenService TokenService, repository repository.Repository) *TelegramImpl {
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
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(impl.ReceiveTelegramMessage),
	}

	b, err := bot.New(cfg.TelegramToken, opts...)
	if err != nil {
		logger.Fatalw("error in creating telegram bot", "error", err)
	}

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
			impl.SendTelegramMessage(update.Message.Chat.ID, "Error in generating token")
			return
		}
		emailBody := fmt.Sprintf("Please click on the below link to verify your email. \n\n %s%s?token=%s", impl.cfg.BaseUrl, util.VerifyTelegramEmail, url.QueryEscape(token))
		err = impl.mailService.SendMail(emails[0], "Get-Link - Email Verification", emailBody)
		if err != nil {
			impl.logger.Errorw("Error in sending mail", "Error", err)
			impl.SendTelegramMessage(update.Message.Chat.ID, "Error in sending verification mail")
			return
		}
		impl.SendTelegramMessage(update.Message.Chat.ID, "Please check your email for verification link")

	} else if update.Message.Text != "" {
		impl.SendTelegramMessage(update.Message.Chat.ID, update.Message.Text)
	} else {

	}
}

func (impl *TelegramImpl) SendTelegramMessage(chatID int64, message string) {
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
	err = impl.repository.InsertUpdateTelegramNumber(encryptedEmail, senderId, chatId)
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
		senderId, err = cryptography.EncryptData(email, senderId, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting data", "Error: ", err)
			return err
		}
	}
	if claims.ChatId != 0 {
		chatId, err = cryptography.EncryptData(email, chatId, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in encrypting data", "Error: ", err)
			return err
		}
	}
	claims.Email = encryptedEmail
	err = impl.repository.InsertUpdateTelegramNumber(encryptedEmail, senderId, chatId)
	if err != nil {
		impl.logger.Errorw("Error in inserting data", "Error: ", err)
		return err
	}

	return nil
}
