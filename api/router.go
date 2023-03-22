package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewAPI(y YoutubeController) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/download", y.DownloadMp3)
	router.HandleFunc("/", y.ServeIndex)
	return router
}
