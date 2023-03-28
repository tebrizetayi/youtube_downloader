package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewAPI(y YoutubeConvertorController) http.Handler {
	router := mux.NewRouter()

	// Create a file server handler that serves static files from the "static" directory
	fileServer := http.FileServer(http.Dir("static"))

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
	// Register the file server handler to serve static files at the "/static/" URL path
	//router.Handle("/static/", http.StripPrefix("/static/", fileServer))
	router.HandleFunc("/download", y.DownloadMp3)
	router.HandleFunc("/", y.ServeIndex)
	return router
}
