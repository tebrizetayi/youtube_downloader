package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func NewAPI(y YoutubeConvertorController) http.Handler {
	router := mux.NewRouter()

	// Add the enforceHTTPS middleware
	router.Use(enforceHTTPS)

	// Create a file server handler that serves static files from the "static" directory
	fileServer := http.FileServer(http.Dir("static"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	// Register the file server handler to serve static files at the "/static/" URL path
	router.HandleFunc("/download", y.DownloadMp3)
	router.HandleFunc("/info", y.Info)
	router.HandleFunc("/", y.ServeIndex)
	router.HandleFunc("/progress", y.ProgressHandler)
	router.HandleFunc("/downloadFile", y.DownloadResultHandler)
	router.HandleFunc("/watch", y.WatchHandler)
	return router
}

func enforceHTTPS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwardedProto := r.Header.Get("X-Forwarded-Proto")
		if strings.ToLower(forwardedProto) != "https" {
			http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, r)
	})
}
