package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"go.uber.org/zap"
)

func getEncryptionKey(encryptionKey string) string {
	key := encryptionKey
	for len(key) < 32 {
		key += encryptionKey
	}
	return key[:32]
}

func getIV(encryptionKey string) string {
	iv := encryptionKey
	for len(iv) < 16 {
		iv += encryptionKey
	}
	return iv[:16]
}

func PKCS5UnPadding(src []byte) []byte {
	length := len(src)
	unPadding := int(src[length-1])
	return src[:(length - unPadding)]
}

func EncryptData(encryptionKey string, plaintext string, logger *zap.SugaredLogger) (string, error) {
	key := getEncryptionKey(encryptionKey)
	iv := getIV(encryptionKey)

	var plainTextBlock []byte
	length := len(plaintext)

	if length%16 != 0 {
		extendBlock := 16 - (length % 16)
		plainTextBlock = make([]byte, length+extendBlock)
		copy(plainTextBlock[length:], bytes.Repeat([]byte{uint8(extendBlock)}, extendBlock))
	} else {
		plainTextBlock = make([]byte, length)
	}

	copy(plainTextBlock, plaintext)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		logger.Errorw("Error in creating cipher", "Error: ", err)
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	ciphertext := make([]byte, len(plainTextBlock))
	mode.CryptBlocks(ciphertext, plainTextBlock)
	str := base64.StdEncoding.EncodeToString(ciphertext)

	return str, nil
}

func DecryptData(encryptionKey string, encryptedString string, logger *zap.SugaredLogger) (string, error) {
	key := getEncryptionKey(encryptionKey)
	iv := getIV(encryptionKey)

	cipherText, err := base64.StdEncoding.DecodeString(encryptedString)
	if err != nil {

	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		logger.Errorw("Error in creating cipher", "Error: ", err)
		return "", err
	}

	if len(cipherText)%aes.BlockSize != 0 {
		return "", fmt.Errorf("block size cant be zero")
	}

	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(cipherText, cipherText)
	decryptedText := PKCS5UnPadding(cipherText)

	return string(decryptedText), nil
}
