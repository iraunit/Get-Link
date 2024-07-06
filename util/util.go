package util

import (
	"encoding/base64"
	"fmt"
	"mime"
	"time"
)

func GetFileExtension(mimeType string) (string, error) {
	extensions, err := mime.ExtensionsByType(mimeType)
	if err != nil {
		return "", err
	}
	if len(extensions) == 0 {
		return "", fmt.Errorf("no extensions found for MIME type %s", mimeType)
	}
	return extensions[0], nil
}

func GetMimeTypeFromExtension(extension string) (string, error) {
	mimeType := mime.TypeByExtension(extension)
	if mimeType == "" {
		return "", fmt.Errorf("no MIME type found for extension %s", extension)
	}
	return mimeType, nil
}

func EncodeString(input string) string {
	return base64.URLEncoding.EncodeToString([]byte(input))
}

func DecodeString(input string) (string, error) {
	decodedBytes, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

func GetFileNameFromType(fileType, mimeType string) string {
	return fmt.Sprintf("%s_%s_From-Get-Link", fileType, time.Now().UTC().Format(time.RFC822))
}
