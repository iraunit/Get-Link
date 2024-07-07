package services

import (
	"github.com/iraunit/get-link-backend/pkg/fileManager"
	"github.com/iraunit/get-link-backend/pkg/repository"
	"github.com/iraunit/get-link-backend/util"
	"go.uber.org/zap"
)

type FileService interface {
	CleanGetLinkAppFiles(email string)
}

type FileServiceImpl struct {
	logger      *zap.SugaredLogger
	repository  repository.Repository
	fileManager fileManager.FileManager
}

func NewFileServiceImpl(logger *zap.SugaredLogger, repository repository.Repository, fileManager fileManager.FileManager) *FileServiceImpl {
	return &FileServiceImpl{
		logger:      logger,
		repository:  repository,
		fileManager: fileManager,
	}
}

func (impl *FileServiceImpl) CleanGetLinkAppFiles(email string) {
	folderPath := impl.fileManager.GetPathToSaveFileFromGetLink(util.EncodeString(email))
	impl.fileManager.DeleteFileFromPathOlderThan24Hours(folderPath)
	folderSize, err := impl.fileManager.GetSizeOfADirectory(folderPath)
	if err != nil {
		impl.logger.Errorw("Error in getting folder size", "Error", err)
	}
	maxLimit := util.FreeGetLinkFileLimitSizeMB

	if folderSize > int64(maxLimit) {
		impl.fileManager.DeleteAllFileFromPath(folderPath)
	}
}
