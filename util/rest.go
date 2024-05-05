package util

import (
	"github.com/caarlos0/env"
	"github.com/go-resty/resty/v2"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
)

type RestClient interface {
	SendWhatsappMessage(url string, body interface{}) (string, error)
}

type RestClientImpl struct {
	logger *zap.SugaredLogger
	cfg    bean.WhatsAppConfig
}

func NewRestClientImpl(logger *zap.SugaredLogger) *RestClientImpl {
	cfg := bean.WhatsAppConfig{}
	if err := env.Parse(&cfg); err != nil {
		return &RestClientImpl{
			logger: logger,
			cfg:    cfg,
		}
	}
	return &RestClientImpl{
		logger: logger,
		cfg:    cfg,
	}
}

func (impl *RestClientImpl) SendWhatsappMessage(url string, body interface{}) (string, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+impl.cfg.AuthToken).
		SetBody(body).
		Post(url)

	if err != nil {
		impl.logger.Errorw("Error in sending whatsapp message", "Error", err)
		return "", err
	}

	return string(resp.Body()), nil
}
