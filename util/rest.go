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
	DownloadMediaFromUrl(url string, token string, filePath string, userEmail string)
}

type RestClientImpl struct {
	logger *zap.SugaredLogger
	cfg    bean.WhatsAppConfig
	async  *Async
}

func NewRestClientImpl(logger *zap.SugaredLogger, async *Async) *RestClientImpl {
	cfg := bean.WhatsAppConfig{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Error in parsing config", "Error", err)
	}
	return &RestClientImpl{
		logger: logger,
		cfg:    cfg,
		async:  async,
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

func (impl *RestClientImpl) DownloadMediaFromUrl(url, token, filePath, userEmail string) {

	impl.async.Run(func() {
		client := resty.New()

		resp, err := client.R().SetHeader("Authorization", "Bearer "+token).Get(url)
		if err != nil {
			impl.logger.Errorw("Error in downloading whatsapp media", "Error", err)
			return
		}

		out, err := os.Create(filePath)
		if err != nil {
			impl.logger.Errorw("Error in creating file", "Error", err)
			return
		}

		defer func(out *os.File) {
			err := out.Close()
			if err != nil {
				impl.logger.Errorw("Error in closing file", "Error", err)
				return
			}
		}(out)

		key, err := CreateKey(userEmail)
		if err != nil {
			impl.logger.Errorw("Error creating encryption key", "Error", err)
			return
		}

		err = EncryptDataAndSaveToFile(out, key, resp.Body(), impl.logger)
		if err != nil {
			impl.logger.Errorw("Error encrypting and saving to file", "Error", err)
			return
		}
	})

}
