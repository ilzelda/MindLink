package page

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/google/uuid"

	"mindlink.io/mindlink/pkg/models"
)

type handler struct {
	log logr.Logger
}

func NewHandler(log logr.Logger) *handler {
	return &handler{
		log: log,
	}
}

func (ph *handler) RegsistRoute(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users/{user_id}/pages", ph.createUserPage)
	mux.HandleFunc("GET /api/users/{user_id}/pages", ph.listUserPages)
	mux.HandleFunc("GET /api/users/{user_id}/pages/{page_id}", ph.loadUserPage)
	mux.HandleFunc("PUT /api/users/{user_id}/pages/{page_id}", ph.saveUserPage)
}

const dataDir string = "data"

func (ph *handler) createUserPage(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	page := &models.Page{
		ID: uuid.New(),
	}
	if err := json.NewDecoder(r.Body).Decode(page); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()
	ph.log.Info("Create user page", "User", userID, "Page", page)

	// 데이터를 저장할 디렉토리 있으면 사용 없으면 생성하여 사용
	dirPath := filepath.Join(dataDir, userID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		http.Error(w, "failed to create data directory", http.StatusInternalServerError)
		return
	}

	// 파일 생성 및 데이터 저장
	fileName := filepath.Join(dirPath, page.ID.String()+".json")
	file, err := os.Create(fileName)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(page); err != nil {
		http.Error(w, "failed to save data", http.StatusInternalServerError)
		return
	}

	response := respUserPage{
		ID:   page.ID,
		Name: page.Name,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (ph *handler) listUserPages(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	userDirPath := filepath.Join(dataDir, userID)
	userPageNames, err := os.ReadDir(userDirPath)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	ph.log.Info("List user pages", "User", userID)

	var respBody respListUserPages = make([]respUserPage, 0, len(userPageNames))
	// TODO: 최근 변경 시간 순으로 정렬
	for _, page := range userPageNames {
		pageFilePath := filepath.Join(userDirPath, page.Name())
		f, err := os.Open(pageFilePath)
		if err != nil {
			http.Error(w, "failed to open file", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		var r respUserPage
		if err := json.NewDecoder(f).Decode(&r); err != nil {
			http.Error(w, "failed to decode", http.StatusInternalServerError)
			return
		}
		respBody = append(respBody, r)
	}

	w.Header().Set("Content-Type", "application/json")
	if len(respBody) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := json.NewEncoder(w).Encode(&respBody); err != nil {
		http.Error(w, "failed to decode", http.StatusInternalServerError)
		return
	}
}

func (ph *handler) loadUserPage(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	pageID := r.PathValue("page_id")
	ph.log.Info("Load user page", "User", userID, "PageID", pageID)

	filePath := filepath.Join(dataDir, userID, pageID+".json")
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file is not exist", http.StatusNotFound)
		} else {
			http.Error(w, "failed to read file", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(fileContent)
}

func (ph *handler) saveUserPage(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	pageID := r.PathValue("page_id")
	ph.log.Info("Save user page", "User", userID, "PageID", pageID)

	pageFilePath := filepath.Join(dataDir, userID, pageID+".json")
	f, err := os.Open(pageFilePath)
	if err != nil {
		http.Error(w, "failed to open file", http.StatusBadRequest)
		return
	}
	defer f.Close()

	var page models.Page
	if err := json.NewDecoder(f).Decode(&page); err != nil {
		http.Error(w, "failed to decode old file data", http.StatusInternalServerError)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&page); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	f, err = os.OpenFile(pageFilePath, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		http.Error(w, "failed to open file", http.StatusBadRequest)
		return
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(&page); err != nil {
		http.Error(w, "failed to encode new file data", http.StatusInternalServerError)
		return
	}

	response := respSaveUserPage{
		SuccessOK: true,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
