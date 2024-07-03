package util

import (
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
