package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func NewAPI(y YoutubeConvertorController) http.Handler {
	router := mux.NewRouter()

	//router.Use(redirectToWWW)
	//router.Use(addCacheControl)

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

	router.Path("/Robots.txt").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})

	router.Path("/Sitemap.xml").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/sitemap.xml")
	})

	router.Path("/es").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index_es.html")
	})

	router.Path("/en").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	router.Path("/de").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index_de.html")
	})

	router.Path("/fr").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index_fr.html")
	})

	router.Path("/ru").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index_ru.html")
	})

	router.Path("/tr").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index_tr.html")
	})

	router.Path("/youtube-mp3").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index_youtube_mp3.html")
	})

	router.Path("/fast-free-easy").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/fast-free-easy.html")
	})

	router.Path("/all-platforms-supported").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/all-platforms-supported.html")
	})

	router.Path("/many-formats-coming-soon").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/many-formats-coming-soon.html")
	})

	router.Path("/without-limitations").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/without-limitations.html")
	})

	router.Path("/safe-and-clean").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/safe-and-clean.html")
	})

	router.Path("/always-up-to-date").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/always-up-to-date.html")
	})

	return router
}

func redirectToWWW(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Printf("Host is %s\n", r)
		host := r.Host
		if strings.HasPrefix(host, "m.m3youtube.com") {
			http.Redirect(w, r, "https://www.m3youtube.com"+r.URL.String(), http.StatusMovedPermanently)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func addCacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=31536000")
		next.ServeHTTP(w, r)
	})
}
