package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"net/http"
	"time"
	"youtube_download/pkg/mp3downloader"
	"youtube_download/pkg/youtubevideoprofiler"
)

type YoutubeConvertorController struct {
	mp3downloader.Mp3downloader
	YVideoprofiler youtubevideoprofiler.ProfilerClient
}

func NewYoutubeController(mp3Downloader mp3downloader.Mp3downloader,
	youtubevideoprofiler youtubevideoprofiler.ProfilerClient,
) YoutubeConvertorController {
	return YoutubeConvertorController{
		mp3Downloader,
		youtubevideoprofiler}

}

// DownloadMp3
func (c *YoutubeConvertorController) DownloadMp3(w http.ResponseWriter, r *http.Request) {
	// Open the file to be downloaded
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	url := r.FormValue("url")
	log.Println("Downloading:", url)

	// Check if the video is exists and public
	IsAvailable, err := c.YVideoprofiler.IsVideoAvailable(ctx, url)
	if !IsAvailable {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the video is longer than 10 minutes
	isValid, _ := c.YVideoprofiler.CheckVideoDuration(ctx, url, 600)
	if !isValid {
		http.Error(w, "Video is longer than 10 minutes", http.StatusBadRequest)
		return
	}

	resultChan := make(chan convertMp3Result)
	go c.ConvertMp3Async(ctx, url, resultChan)

	var result convertMp3Result
	select {
	case result = <-resultChan:
		if result.Err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)

			return
		}
	case <-ctx.Done():
		log.Println("Connection is closed")
		return
	}

	// Set the response header to indicate the file download
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("%s.mp3", time.Now().String()))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(result.Result)))
	log.Println("Downloaded:", url)
	w.Write(result.Result)
}

// ServeIndex
func (c *YoutubeConvertorController) ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

// Get url and response the time of the request time
func (c *YoutubeConvertorController) GetTime(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(time.Now().String()))
}

type VideoInfo struct {
	Duration string `json:"duration"`
	Title    string `json:"title"`
	Error    string `json:"error"`
}

func (c *YoutubeConvertorController) Info(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	url := r.URL.Query().Get("url")

	video, err := c.YVideoprofiler.GetVideoInfo(ctx, url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	videoInfo := VideoInfo{
		Duration: fmt.Sprintf("%d:%d minutes",
			time.Duration(video.Duration.Seconds())/60,
			time.Duration(video.Duration.Seconds())%60),
		Title: video.Title,
	}
	jsonData, _ := json.Marshal(videoInfo)

	w.Write(jsonData)
}

type convertMp3Result struct {
	Result []byte
	Err    error
}

func (c *YoutubeConvertorController) ConvertMp3Async(ctx context.Context, url string, result chan<- convertMp3Result) {
	// Create a channel to receive the result of DownloadMp3
	downloadResult := make(chan convertMp3Result)

	// Call DownloadMp3 in a goroutine
	go func() {
		mp3File, _, err := c.Mp3downloader.DownloadMp3(ctx, url)
		downloadResult <- convertMp3Result{
			Result: mp3File,
			Err:    err,
		}
	}()

	// Wait for either the completion of DownloadMp3 or the cancellation of the context
	select {
	case res := <-downloadResult:
		result <- res
	case <-ctx.Done():
		result <- convertMp3Result{
			Err: ctx.Err(),
		}
	}
}

/*
func (c *YoutubeConvertorController) DownloadMp3(w http.ResponseWriter, r *http.Request) {
	// Open the file to be downloaded
	ctx := r.Context()
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	url := r.FormValue("url")
	log.Println("Downloading:", url)

	// Check if the video is exists and public
	IsAvailable, err := c.YVideoprofiler.IsAvailable(ctx, url)
	if !IsAvailable {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the video is longer than 10 minutes
	isValid, _ := c.YVideoprofiler.CheckDuration(ctx, url, 600)
	if !isValid {
		http.Error(w, "Video is longer than 10 minutes", http.StatusBadRequest)
		return
	}

	mp3FileName, err := c.Mp3downloader.ConvertorMp3(ctx, url)
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
}*/
