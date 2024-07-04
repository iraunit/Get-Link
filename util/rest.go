package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-resty/resty/v2"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/http"
	"os"
)

type RestClient interface {
	SendWhatsappMessage(url string, body interface{}) (string, error)
	GetMediaDataFromId(url string) (*bean.WhatsappMedia, error)
	DownloadMediaFromUrl(url string, token string, filePath string) error
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

func (impl *RestClientImpl) GetMediaDataFromId(url string) (*bean.WhatsappMedia, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+impl.cfg.AuthToken).
		Get(url)

	if err != nil {
		impl.logger.Errorw("Error in getting whatsapp media", "Error", err)
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		impl.logger.Errorw("Error in getting whatsapp message. status not ok.", "Error", err)
		return nil, errors.New(fmt.Sprintf("Status : %d", resp.StatusCode()))
	}

	var media bean.WhatsappMedia
	err = json.Unmarshal(resp.Body(), &media)
	if err != nil {
		impl.logger.Errorw("Error unmarshalling response body", "Error", err)
		return nil, err
	}

	return &media, nil
}

func (impl *RestClientImpl) DownloadMediaFromUrl(url string, token string, filePath string) error {
	client := resty.New()

	resp, err := client.R().SetHeader("Authorization", "Bearer "+token).Get(url)
	if err != nil {
		impl.logger.Errorw("Error in downloading whatsapp media", "Error", err)
		return err
	}

	out, err := os.Create(filePath)
	if err != nil {
		impl.logger.Errorw("Error in creating file", "Error", err)
		return err
	}

	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			impl.logger.Errorw("Error in closing file", "Error", err)
			return
		}
	}(out)

	_, err = out.Write(resp.Body())
	if err != nil {
		impl.logger.Errorw("Error in writing to file", "Error", err)
		return err
	}

	return nil
}
