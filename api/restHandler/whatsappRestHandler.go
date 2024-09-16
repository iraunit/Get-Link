package restHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-pg/pg"
	"github.com/gorilla/context"
	"github.com/iraunit/get-link-backend/pkg/services"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Whatsapp interface {
	Verify(w http.ResponseWriter, r *http.Request)
	HandleMessage(w http.ResponseWriter, r *http.Request)
	SendWhatsappMessage(w http.ResponseWriter, r *http.Request)
}

type WhatsappImpl struct {
	logger   *zap.SugaredLogger
	cfg      bean.WhatsAppConfig
	wService services.WhatsappService
}

func NewWhatsappImpl(logger *zap.SugaredLogger, wService services.WhatsappService) *WhatsappImpl {
	cfg := bean.WhatsAppConfig{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error loading Cfg from env", "Error", zap.Error(err))
	}
	return &WhatsappImpl{
		logger:   logger,
		cfg:      cfg,
		wService: wService,
	}
}

func (impl *WhatsappImpl) Verify(w http.ResponseWriter, r *http.Request) {
	verifyToken := r.URL.Query().Get("hub.verify_token")
	if verifyToken == impl.cfg.VerifyToken {
		impl.logger.Infow("Webhook verification successful", "verifyToken", verifyToken)
		_, err := w.Write([]byte(r.URL.Query().Get("hub.challenge")))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			impl.logger.Errorw("Error writing response", "Error: ", err)
			_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 500, Error: "Error in writing response"})
			return
		}
		return
	}
	impl.logger.Infow("Webhook verification failed", "verifyToken", verifyToken)
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte("Invalid verification token"))
	return
}

func (impl *WhatsappImpl) HandleMessage(w http.ResponseWriter, r *http.Request) {

	message := bean.WhatsAppBusinessMessage{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		impl.logger.Errorw("Error reading request body", "Error: ", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		impl.logger.Errorw("Error in decoding request body", "Body", string(body), "Error: ", err)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 500, Error: "Error in decoding request body"})
		return
	}

	for i := 0; i < len(message.Entry); i++ {
		for j := 0; j < len(message.Entry[i].Changes); j++ {
			currentMessages := message.Entry[i].Changes[j].Value.Messages
			for k := 0; k < len(currentMessages); k++ {
				impl.logger.Infow("Message Received", "From", currentMessages[k].From)
				err = impl.wService.ReceiveMessage(&currentMessages[k])
				if err != nil {
					if errors.Is(err, pg.ErrNoRows) {
						_ = impl.wService.SendMessage(message.Entry[i].Changes[j].Value.Contacts[0].WaID, fmt.Sprintf("Hey %s!\nPlease set and verify your email first.\nSend \n> *`set email youremail`*\n here to get started. You can share your feedback or report an issue on codingkaro.in.\n\nRegards\nRaunit Verma\nShypt Solution", message.Entry[i].Changes[j].Value.Contacts[0].Profile.Name))
					} else {
						_ = impl.wService.SendMessage(message.Entry[i].Changes[j].Value.Contacts[0].WaID, fmt.Sprintf("Hey %s,\nSorry for the inconvenience. Get-Link is unable to process your request. Please try again later. Please share your feedback or report an issue on codingkaro.in\n\nRegards\nRaunit Verma\nShypt Solution", message.Entry[i].Changes[j].Value.Contacts[0].Profile.Name))
						impl.logger.Errorw("Error in handling message", "Message:", currentMessages[k], "Error: ", err)
					}
				}
			}
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (impl *WhatsappImpl) SendWhatsappMessage(w http.ResponseWriter, r *http.Request) {
	userEmail := context.Get(r, "email").(string)
	allEmails := impl.wService.GetIfUserIsPremium(userEmail)
	if !allEmails {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 403, Error: "You are not a premium user. Please upgrade your plan. Contact me on twitter (iraunit) or email me on contact.shyptsolution@gmail.com."})
		return
	}

	msg := &Message{}
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "Error in decoding request body"})
		return
	}

	err = impl.wService.SendMessageFromWeb(userEmail, msg.Message)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: err.Error()})
		return
	}

	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: "Message sent successfully"})
}
