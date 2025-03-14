package restHandler

import (
	"encoding/json"
	"fmt"
	muxContext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/iraunit/get-link-backend/pkg/fileManager"
	"github.com/iraunit/get-link-backend/pkg/services"
	"github.com/iraunit/get-link-backend/util"
	"github.com/iraunit/get-link-backend/util/bean"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
)

type FileHandler interface {
	DownloadFile(w http.ResponseWriter, r *http.Request)
	UploadFile(w http.ResponseWriter, r *http.Request)
	ListAllFiles(w http.ResponseWriter, r *http.Request)
	DeleteFile(w http.ResponseWriter, r *http.Request)
}

type FileHandlerImpl struct {
	logger        *zap.SugaredLogger
	fileManager   fileManager.FileManager
	fileService   services.FileService
	downloadQueue chan struct{}
}

func NewFileHandlerImpl(logger *zap.SugaredLogger, fileManager fileManager.FileManager, fileService services.FileService) *FileHandlerImpl {
	return &FileHandlerImpl{
		logger:        logger,
		fileManager:   fileManager,
		fileService:   fileService,
		downloadQueue: make(chan struct{}, 1),
	}
}

func (impl *FileHandlerImpl) DownloadFile(w http.ResponseWriter, r *http.Request) {
	impl.downloadQueue <- struct{}{}
	defer func() {
		<-impl.downloadQueue
	}()

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
	email := muxContext.Get(r, "email").(string)

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			impl.logger.Errorw("Error in closing file", "Error", err)
		}
	}(file)

	filename := header.Filename

	impl.fileService.CleanGetLinkAppFiles(email)

	err = impl.fileManager.SaveFileToPath(file, fmt.Sprintf("%s/%s.bin", impl.fileManager.GetPathToSaveFileFromApp(util.EncodeString(email), util.GETLINK), filename), email)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: err.Error()})
		impl.logger.Errorw("Error in uploading file", "Error", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (impl *FileHandlerImpl) ListAllFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "https://codingkaro.in")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, DELETE, GET, OPTIONS, HEAD")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	email := muxContext.Get(r, "email").(string)

	var allFiles []bean.FileInfo
	appNames := []string{util.WHATSAPP, util.GETLINK, util.TELEGRAM}

	for _, appName := range appNames {
		files, err := impl.fileManager.ListAllFilesFromApp(email, appName)
		if err != nil {
			impl.logger.Errorw("Error in listing file", "appName", appName, "email", email, "Error", err)
			continue
		}
		allFiles = append(allFiles, files...)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 200, Result: allFiles})
	impl.fileManager.DeleteAllFileOlderThanHours("/tmp/data", 240)
}

func (impl *FileHandlerImpl) DeleteFile(w http.ResponseWriter, r *http.Request) {
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

	err := impl.fileManager.DeleteAFileInAppFolder(fileName, email, appName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(bean.Response{StatusCode: 400, Error: err.Error()})
		impl.logger.Errorw("Error in deleting file", "Error", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
