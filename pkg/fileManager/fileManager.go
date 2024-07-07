package fileManager

import (
	"fmt"
	"github.com/iraunit/get-link-backend/pkg/cryptography"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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
	ListAllFilesFromApp(userEmail, appName string) ([]bean.FileInfo, error)
	SaveFileToPath(data io.ReadCloser, path, userEmail string) error
	DeleteAllFileOlderThanHours(path string, hours int)
	DeleteAFileInAppFolder(fileName, email, app string) error
}

type FileManagerImpl struct {
	logger *zap.SugaredLogger
	async  *util.Async
}

func NewFileManagerImpl(logger *zap.SugaredLogger, async *util.Async) *FileManagerImpl {
	return &FileManagerImpl{
		logger: logger,
		async:  async,
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

	key, err := cryptography.CreateKey(email)
	if err != nil {
		impl.logger.Errorw("Error creating decryption key", "Error", err)
		return err
	}

	fileName := path.Base(encryptedFilePath)
	w.Header().Set("Content-Disposition", "attachment; filename="+strings.TrimPrefix(fileName, ".bin"))

	mimeType, err := util.GetMimeTypeFromExtension(fmt.Sprintf(".%s", strings.Split(strings.TrimPrefix(fileName, ".bin"), ".")[len(strings.Split(fileName, "."))-2]))
	if err != nil {
		impl.logger.Errorw("Error in getting mime type", "Error", err)
	} else {
		w.Header().Set("Content-Type", mimeType)
	}

	err = cryptography.DecryptFileAndSend(w, key, encryptedData, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error decrypting data", "Error", err)
		return err
	}

	impl.logger.Infow("File successfully decrypted and sent")
	return nil
}

func (impl *FileManagerImpl) ListAllFilesFromApp(userEmail, appName string) ([]bean.FileInfo, error) {
	var allFiles []bean.FileInfo
	folderPath := fmt.Sprintf(util.PathToFiles, userEmail, appName)
	files, err := os.ReadDir(folderPath)
	if err != nil {
		impl.logger.Errorw("Error in reading directory", "Error", err)
		return nil, err
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			impl.logger.Errorw("Error in getting file info", "Error", err)
			continue
		}
		mimeType, err := util.GetMimeTypeFromExtension(fmt.Sprintf(".%s", strings.Split(strings.TrimPrefix(file.Name(), ".bin"), ".")[len(strings.Split(file.Name(), "."))-2]))
		if err != nil {
			impl.logger.Errorw("Error in getting mime type", "Error", err)
			continue
		}
		allFiles = append(allFiles, bean.FileInfo{Name: strings.TrimSuffix(info.Name(), ".bin"), Size: info.Size(), ModTime: info.ModTime(), MimeType: mimeType, AppName: appName})
	}
	return allFiles, nil
}

func (impl *FileManagerImpl) SaveFileToPath(data io.ReadCloser, path, userEmail string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		impl.logger.Errorw("Error in creating directories", "Error", err)
		return err
	}

	out, err := os.Create(path)
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

	key, err := cryptography.CreateKey(userEmail)
	if err != nil {
		impl.logger.Errorw("Error creating encryption key", "Error", err)
		return err
	}

	err = cryptography.EncryptDataAndSaveToFile(out, key, data, impl.logger)
	if err != nil {
		impl.logger.Errorw("Error encrypting and saving to file", "Error", err)
		return err
	}
	return nil
}

func (impl *FileManagerImpl) DeleteAllFileOlderThanHours(p string, hours int) {
	impl.async.Run(func() {
		files, err := os.ReadDir(p)
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
			if info.IsDir() {
				impl.DeleteAllFileOlderThanHours(path.Join(p, info.Name()), hours)
			} else {
				if time.Since(info.ModTime()).Hours() > float64(hours) {
					err = os.Remove(path.Join(p, info.Name()))
					if err != nil {
						impl.logger.Errorw("Error in deleting file", "Error", err)
						continue
					}
					impl.logger.Infow("File deleted", "File", info.Name())
				}
			}
		}
	})
}

func (impl *FileManagerImpl) DeleteAFileInAppFolder(fileName, email, app string) error {
	p := fmt.Sprintf(util.PathToFiles, util.EncodeString(email), app)
	files, err := os.ReadDir(p)
	if err != nil {
		impl.logger.Errorw("Error in reading directory", "Error", err)
		return err
	}

	for _, file := range files {
		if file.Name() == fileName {
			err = os.Remove(path.Join(p, fileName))
			if err != nil {
				impl.logger.Errorw("Error in deleting file", "Error", err)
				return err
			}
			return nil
		}
	}
	return nil
}
