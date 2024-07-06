package cryptography

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
	"io"
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
	if len(src) == 0 {
		return nil
	}
	padding := int(src[len(src)-1])
	if padding >= len(src) {
		return nil
	}
	for i := len(src) - 1; i > len(src)-padding-1; i-- {
		if int(src[i]) != padding {
			return nil
		}
	}
	return src[:len(src)-padding]
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
		logger.Errorw("Error in decoding string", "Error: ", err)
		return "", err
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

func EncryptDataAndSaveToFile(w io.Writer, key []byte, data io.ReadCloser, logger *zap.SugaredLogger) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Errorw("Error creating cipher block", "Error", err)
		return err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		logger.Errorw("Error generating IV", "Error", err)
		return err
	}

	if _, err := w.Write(iv); err != nil {
		logger.Errorw("Error writing IV to file", "Error", err)
		return err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: w}

	if _, err := io.Copy(writer, data); err != nil {
		logger.Errorw("Error writing encrypted data to file", "Error", err)
		return err
	}
	//stream.XORKeyStream(data, data)
	//
	//if _, err := w.Write(data); err != nil {
	//	logger.Errorw("Error writing encrypted data to file", "Error", err)
	//	return err
	//}

	return nil
}

func CreateKey(email string) ([]byte, error) {
	salt := []byte(email)
	key := argon2.IDKey([]byte(email), salt, 1, 64*1024, 4, 32)
	return key, nil
}

func DecryptFileAndSend(w io.Writer, key []byte, data []byte, logger *zap.SugaredLogger) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Errorw("Error creating cipher block", "Error", err)
		return err
	}

	if len(data) < aes.BlockSize {
		return fmt.Errorf("ciphertext too short")
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	if _, err := w.Write(data); err != nil {
		logger.Errorw("Error writing decrypted data to file", "Error", err)
		return err
	}

	return nil
}
