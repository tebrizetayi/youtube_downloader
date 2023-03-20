package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewAPI(y YoutubeController) http.Handler {
	router := mux.NewRouter()
	http.HandleFunc("/download", y.DownloadMp3)
	http.HandleFunc("/", y.ServeIndex)
	return router
}
