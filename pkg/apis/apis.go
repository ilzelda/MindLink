package apis

import (
	"net/http"

	"mindlink.io/mindlink/pkg/apis/internal/auth"
	"mindlink.io/mindlink/pkg/apis/internal/page"
	prepo "mindlink.io/mindlink/pkg/apis/internal/page/repository"
	"mindlink.io/mindlink/pkg/apis/internal/user"
	urepo "mindlink.io/mindlink/pkg/apis/internal/user/repository"
	"mindlink.io/mindlink/pkg/log"
)

const (
	pageFSRoot = "data/page"
	userFSRoot = "data/user"
)

func SetupAPIs() (*http.ServeMux, error) {
	mux := http.NewServeMux()

	pageLogger := log.Logger.WithName("PageAPI")
	pageAPI := page.NewHandler(
		pageLogger,
		prepo.NewFileRepo(pageFSRoot, pageLogger),
		auth.HeaderHandler,
	)
	pageAPI.RegistRoute(mux)

	userLogger := log.Logger.WithName("UserAPI")
	userAPI := user.NewHandler(userLogger)
	userAPI.RegistRoute(mux)

	authAPI, err := auth.NewHandler(
		log.Logger.WithName("AuthAPI"),
		user.NewUsecase(userLogger, urepo.NewFileRepo(userFSRoot, userLogger)),
	)
	if err != nil {
		return nil, err
	}
	authAPI.RegistRoute(mux)

	return mux, nil
}
