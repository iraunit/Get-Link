package restHandler

import (
	"encoding/json"
	"fmt"
	muxContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/iraunit/get-link-backend/pkg/fileManager"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"net/http"
)

type FileHandler interface {
	DownloadFile(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
}

type FileHandlerImpl struct {
	logger      *zap.SugaredLogger
	fileManager fileManager.FileManager
}

func NewFileHandlerImpl(logger *zap.SugaredLogger, fileManager fileManager.FileManager) *FileHandlerImpl {
	return &FileHandlerImpl{
		logger:      logger,
		fileManager: fileManager,
	}
}

func (impl *FileHandlerImpl) DownloadFile(w http.ResponseWriter, r *http.Request) {
	email := muxContext.Get(r, "email").(string)
	vars := mux.Vars(r)
	fileName := vars["fileName"] + ".bin"
	appName := vars["appName"]

	if fileName == "" || appName == "" || email == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: "fileName or appName is missing"})
		impl.logger.Errorw("fileName or appName is missing")
		return
	}

	err := impl.fileManager.DownloadDecryptedFile(w, fmt.Sprintf("%s/%s", impl.fileManager.GetPathToSaveFileFromApp(util.EncodeString(email), appName), fileName), email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: err.Error()})
		impl.logger.Errorw("Error in downloading file", "Error", err)
		return
	}
}

func (impl *FileHandlerImpl) UploadFile(w http.ResponseWriter, r *http.Request) {

}
