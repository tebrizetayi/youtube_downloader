package api

import (
	"fmt"
	"log"

	"net/http"
	"time"
	"youtube_download/mp3downloader"
	"youtube_download/youtubevideoprofiler"
)

type YoutubeConvertorController struct {
	mp3downloader.Mp3downloader
	YVideoprofiler youtubevideoprofiler.YVideoprofiler
}

func NewYoutubeController(mp3Downloader mp3downloader.Mp3downloader,
	youtubevideoprofiler youtubevideoprofiler.YVideoprofiler,
) YoutubeConvertorController {
	return YoutubeConvertorController{
		mp3Downloader,
		youtubevideoprofiler}

}

// DownloadMp3
func (c *YoutubeConvertorController) DownloadMp3(w http.ResponseWriter, r *http.Request) {
	// Open the file to be downloaded

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	url := r.FormValue("url")
	log.Println("Downloading:", url)

	// Check if the video is longer than 10 minutes
	isValid, _ := c.YVideoprofiler.CheckDuration(url, 600)
	if !isValid {
		http.Error(w, "Video is longer than 10 minutes", http.StatusBadRequest)
		return
	}

	mp3File, _, err := c.Mp3downloader.DownloadMp3(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Set the response header to indicate the file download
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("%s.mp3", time.Now().String()))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(mp3File)))

	// Copy the file to the response writer

	log.Println("Downloaded:", url)
	w.Write(mp3File)
}

// ServeIndex
func (c *YoutubeConvertorController) ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

// Get url and response the time of the request time
func (c *YoutubeConvertorController) GetTime(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}
