package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

	fileName, err := c.YoutubeDownloader.DownloadYouTubeMP3(url)
	log.Println(fileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	log.Println(file.Name())
	log.Println(fileInfo)
	// Set the response header to indicate the file download
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Name())
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Copy the file to the response writer
	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Failed to download file", http.StatusInternalServerError)
		return
	}
}

// ServeIndex
func (c *YoutubeController) ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}
