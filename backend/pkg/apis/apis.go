package apis

import (
	"net/http"

	"mindlink.io/mindlink/pkg/apis/page"
	"mindlink.io/mindlink/pkg/log"
	"mindlink.io/mindlink/pkg/repository"
)

const pageFSRoot = "data"

type Handler interface {
	RegsistRoute(*http.ServeMux)
}

var PageAPI Handler = func() Handler {
	pageLogger := log.Logger.WithName("PageAPI")

	return page.NewHandler(
		pageLogger,
		repository.NewPageFSRepo(pageFSRoot, pageLogger),
	)
}()
