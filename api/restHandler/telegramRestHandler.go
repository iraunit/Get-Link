package restHandler

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/context"
	"github.com/iraunit/get-link-backend/pkg/cryptography"
	"github.com/iraunit/get-link-backend/pkg/services"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
)

type TelegramRestHandler interface {
	VerifyTelegramEmail(w http.ResponseWriter, r *http.Request)
	SendTelegramMessage(w http.ResponseWriter, r *http.Request)
}

type TelegramRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	telegramService services.TelegramService
	cfg             *bean.TelegramCfg
}

func NewTelegramRestHandler(logger *zap.SugaredLogger, telegramService services.TelegramService) *TelegramRestHandlerImpl {
	cfg := &bean.TelegramCfg{}
	if err := env.Parse(cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}
	return &TelegramRestHandlerImpl{
		logger:          logger,
		telegramService: telegramService,
		cfg:             cfg,
	}
}

func (impl *TelegramRestHandlerImpl) VerifyTelegramEmail(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	tokenStr, _ := url.QueryUnescape(query.Get("token"))

	claims := bean.TelegramVerificationClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(impl.cfg.JwtKey), nil
	})

	if err != nil {
		impl.logger.Errorw("Unauthorised Request. Invalid token.")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error parsing token."})
		return
	}

	err = impl.telegramService.VerifyTelegram(claims.Email, &claims)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error in verifying link"})
		return
	}
	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: "Link verified successfully"})
}

type Message struct {
	Message string `json:"message"`
}

func (impl *TelegramRestHandlerImpl) SendTelegramMessage(w http.ResponseWriter, r *http.Request) {
	userEmail := context.Get(r, "email").(string)
	allEmails, err := impl.telegramService.GetUsersFromEmail(userEmail)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error in sending message"})
		return
	}

	msg := &Message{}
	err = json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error in decoding request body"})
		return
	}

	for _, email := range allEmails {
		decryptedData, err := cryptography.DecryptData(userEmail, email.ChatId, impl.logger)
		if err != nil {
			impl.logger.Errorw("Error in decrypting data", "Error: ", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		chatId, err := strconv.ParseInt(decryptedData, 10, 64)
		if err != nil {
			impl.logger.Errorw("Error in converting string to int", "Error: ", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		impl.telegramService.SendTelegramMessage(chatId, msg.Message)
	}

	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: "Message sent successfully"})
}
