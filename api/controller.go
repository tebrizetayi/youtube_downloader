package api

import (
	"fmt"
	"log"

	"net/http"
	"time"
	"youtube_download/youtube"
)

type YoutubeController struct {
	youtube.YoutubeDownloader
}

func NewYoutubeController(downloader youtube.YoutubeDownloader) YoutubeController {
	return YoutubeController{downloader}

}

// DownloadMp3
func (c *YoutubeController) DownloadMp3(w http.ResponseWriter, r *http.Request) {
	// Open the file to be downloaded

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	url := r.FormValue("url")
	log.Println("Downloading:", url)

	mp3File, err := c.YoutubeDownloader.DownloadYouTubeMP3(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Set the response header to indicate the file download
	w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("%s.mp3", time.Now().String()))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(mp3File)))

	// Copy the file to the response writer

	log.Println("Downloaded:", url)
	w.Write(mp3File)
}

// ServeIndex
func (c *YoutubeController) ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
