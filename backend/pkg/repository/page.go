package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"mindlink.io/mindlink/pkg/models"
)

type pageFileRepository struct {
	rootDir string
	logger  logr.Logger
}

func NewPageFSRepo(root string, logger logr.Logger) *pageFileRepository {
	return &pageFileRepository{
		rootDir: root,
		logger:  logger,
	}
}

func (repo *pageFileRepository) CreatePage(userID models.UserID, params models.CreatePageParams) (*models.Page, error) {
	// 데이터를 저장할 디렉토리 있으면 사용 없으면 생성하여 사용
	dirPath := filepath.Join(repo.rootDir, userID.String())
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, errors.New("failed to create data directory")
	}

	page := &models.Page{
		ID:            uuid.New(),
		Name:          params.Name,
		NodeNum:       0,
		ConnectionNum: 0,
	}

	// 파일 생성 및 데이터 저장
	fileName := filepath.Join(dirPath, page.ID.String()+".json")
	file, err := os.Create(fileName)
	if err != nil {
		return nil, errors.New("failed to create file")
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(page); err != nil {
		return nil, errors.New("failed to save data")
	}

	return page, nil
}

var (
	ErrUserNotFound       error = errors.New("user not found")
	ErrFailedToAccessPage error = errors.New("failed to access page")
	ErrFailedToDecodePage error = errors.New("failed to decode page")
)

func (repo *pageFileRepository) ListUserPages(userID models.UserID) ([]*models.Page, error) {
	userDirPath := filepath.Join(repo.rootDir, userID.String())
	userPageNames, err := os.ReadDir(userDirPath)
	if err != nil {
		return nil, ErrUserNotFound
	}

	pages := make([]*models.Page, 0, len(userPageNames))
	// TODO: 최근 변경 시간 순으로 정렬
	for _, page := range userPageNames {
		pageFilePath := filepath.Join(userDirPath, page.Name())
		file, err := os.Open(pageFilePath)
		if err != nil {
			return nil, ErrFailedToAccessPage
		}
		defer file.Close()

		var page models.Page
		if err := json.NewDecoder(file).Decode(&page); err != nil {
			return nil, ErrFailedToDecodePage
		}
		pages = append(pages, &page)
	}
	return pages, nil
}

var (
	ErrPageNotFound       error = errors.New("page not found")
	ErrFailedToFindPage   error = errors.New("failed to find page")
	ErrFailedToEncodePage error = errors.New("failed to encode page")
)

func (repo *pageFileRepository) GetPage(userID models.UserID, pageID uuid.UUID) (*models.Page, error) {
	filePath := filepath.Join(repo.rootDir, userID.String(), pageID.String()+".json")
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrPageNotFound
		}
		return nil, ErrFailedToFindPage
	}

	var page models.Page
	if err := json.Unmarshal(fileContent, &page); err != nil {
		return nil, ErrFailedToDecodePage
	}

	return &page, nil
}

func (repo *pageFileRepository) UpdatePage(userID models.UserID, pageID uuid.UUID, params models.UpdatePageParams) (*models.Page, error) {
	pageFilePath := filepath.Join(repo.rootDir, userID.String(), pageID.String()+".json")
	oldFile, err := os.Open(pageFilePath)
	if err != nil {
		return nil, ErrFailedToAccessPage
	}
	defer oldFile.Close()

	var page models.Page
	if err := json.NewDecoder(oldFile).Decode(&page); err != nil {
		return nil, ErrFailedToDecodePage
	}

	if params.Name != "" {
		page.Name = params.Name
	}
	if len(params.Nodes) != 0 {
		page.Nodes = params.Nodes
		page.NodeNum = len(params.Nodes)
	}
	if len(params.Connections) != 0 {
		page.Connections = params.Connections
		page.ConnectionNum = len(params.Connections)
	}

	file, err := os.OpenFile(pageFilePath, os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return nil, ErrFailedToAccessPage
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(&page); err != nil {
		return nil, ErrFailedToEncodePage
	}

	return &page, nil
}

func (repo *pageFileRepository) DeletePage(userID models.UserID, pageID uuid.UUID) (*models.Page, error) {
	filePath := filepath.Join(repo.rootDir, userID.String(), pageID.String()+".json")
	var page models.Page
	if err := func() error {
		oldFile, err := os.Open(filePath)
		if err != nil {
			return ErrFailedToAccessPage
		}
		defer oldFile.Close()

		if err := json.NewDecoder(oldFile).Decode(&page); err != nil {
			return ErrFailedToDecodePage
		}

		return nil
	}(); err != nil {
		return nil, err
	}

	if err := os.Remove(filePath); err != nil {
		return nil, errors.New("failed to delete page")
	}

	return &page, nil
}
