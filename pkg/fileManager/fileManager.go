package fileManager

import (
	"fmt"
	"github.com/iraunit/get-link-backend/util"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

type FileManager interface {
	GetPathToSaveFileFromWhatsapp(userEmail string) string
	GetPathToSaveFileFromTelegram(userEmail string) string
	GetPathToSaveFileFromGetLink(userEmail string) string
	GetPathToSaveFileFromApp(userEmail, appName string) string
	DeleteFileFromPathOlderThan24Hours(path string)
	DeleteAllFileFromPath(path string)
	GetSizeOfADirectory(path string) (int64, error)
	DownloadDecryptedFile(w http.ResponseWriter, encryptedFilePath, email string) error
}

type FileManagerImpl struct {
	logger *zap.SugaredLogger
}

func NewFileManagerImpl(logger *zap.SugaredLogger) *FileManagerImpl {
	return &FileManagerImpl{
		logger: logger,
	}
}

func (impl *FileManagerImpl) GetPathToSaveFileFromWhatsapp(userEmail string) string {
	return fmt.Sprintf(util.PathToFiles, userEmail, util.WHATSAPP)
}

func (impl *FileManagerImpl) GetPathToSaveFileFromTelegram(userEmail string) string {
	return fmt.Sprintf(util.PathToFiles, userEmail, util.TELEGRAM)
}

func (impl *FileManagerImpl) GetPathToSaveFileFromGetLink(userEmail string) string {
	return fmt.Sprintf(util.PathToFiles, userEmail, util.GETLINK)
}

func (impl *FileManagerImpl) GetPathToSaveFileFromApp(userEmail, appName string) string {
	return fmt.Sprintf(util.PathToFiles, userEmail, appName)
}

func (impl *FileManagerImpl) DeleteFileFromPathOlderThan24Hours(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		impl.logger.Errorw("Error in reading directory", "Error", err)
		return
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			impl.logger.Errorw("Error in getting file info", "Error", err)
			continue
		}
		if info.ModTime().Before(time.Now().Add(-24 * time.Hour)) {
			impl.DeleteFileFromPath(fmt.Sprintf("%s/%s", path, info.Name()))
		}
	}
}

func (impl *FileManagerImpl) DeleteAllFileFromPath(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		impl.logger.Errorw("Error in reading directory", "Error", err)
		return
	}

	for _, file := range files {
		impl.DeleteFileFromPath(fmt.Sprintf("%s/%s", path, file.Name()))
	}
}

func (impl *FileManagerImpl) DeleteFileFromPath(path string) {
	err := os.Remove(path)
	if err != nil {
		impl.logger.Errorw("Error in deleting file", "Error", err)
		return
	}
}

func EnsureDirectory(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}

func (impl *FileManagerImpl) GetSizeOfADirectory(path string) (int64, error) {
	err := EnsureDirectory(path)
	if err != nil {
		return 0, err
	}

	d, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func(d *os.File) {
		err := d.Close()
		if err != nil {
			impl.logger.Errorw("Error in closing file", "Error", err)
		}
	}(d)

	stats, err := d.Stat()
	if err != nil {
		return 0, err
	}
	return stats.Size() / 1000000, nil
}

func (impl *FileManagerImpl) DownloadDecryptedFile(w http.ResponseWriter, encryptedFilePath, email string) error {
	encryptedData, err := os.ReadFile(encryptedFilePath)
	if err != nil {
		impl.logger.Errorw("Error reading encrypted file", "Error", err)
		return err
	}

	key, err := util.CreateKey(email)
	if err != nil {
		impl.logger.Errorw("Error creating decryption key", "Error", err)
		return err
	}

	err = util.DecryptFileAndSend(w, key, encryptedData, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error decrypting data", "Error", err)
		return err
	}

	impl.logger.Infow("File successfully decrypted and sent")
	return nil
}
