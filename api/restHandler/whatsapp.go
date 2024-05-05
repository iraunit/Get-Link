package restHandler

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/iraunit/get-link-backend/pkg/services"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type Whatsapp interface {
	Verify(w http.ResponseWriter, r *http.Request)
	HandleMessage(w http.ResponseWriter, r *http.Request)
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
				err = impl.wService.ReceiveMessage(&currentMessages[k])
				if err != nil {
					_ = impl.wService.SendMessage(message.Entry[i].Changes[j].Value.Contacts[0].WaID, fmt.Sprintf("Hey %s, Sorry for the inconvenience. Get-Link am unable to process your request. Please try again later. Please share your feedback or report an issue on codingkaro.in \n\nThank you.", message.Entry[i].Changes[j].Value.Contacts[0].Profile.Name))
					impl.logger.Errorw("Error in handling message", "Message:", currentMessages[k], "Error: ", err)
				}
			}
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
