package util

import (
	"encoding/base64"
	"fmt"
	"mime"
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
