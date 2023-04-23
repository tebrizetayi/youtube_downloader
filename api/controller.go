package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"net/http"
	"time"
	"youtube_download/internal/mp3downloader"
	"youtube_download/internal/youtubevideoprofiler"
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

// DownloadProgress struct for progress tracking
type DownloadProgress struct {
	sync.Mutex
	progress  bool
	lastCheck time.Time
	context.CancelFunc
}

// Store the download progress and result in a map
var downloadProgress = make(map[string]*DownloadProgress)
var downloadResults = make(map[string][]byte)

func (dp *DownloadProgress) SetProgress(progress bool) {
	dp.Lock()
	defer dp.Unlock()
	dp.progress = progress
}

func (dp *DownloadProgress) GetProgress() bool {
	dp.Lock()
	defer dp.Unlock()
	dp.lastCheck = time.Now()
	return dp.progress
}

func (dp *DownloadProgress) GetLastCheck() time.Time {
	dp.Lock()
	defer dp.Unlock()
	return dp.lastCheck
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

	// Generate a unique download token
	downloadToken := url
	progress := &DownloadProgress{}
	downloadProgress[downloadToken] = progress

	// Check if the video is exists and public
	if IsAvailable, err := c.YVideoprofiler.IsVideoAvailable(ctx, url); err != nil || !IsAvailable {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !IsAvailable {
			http.Error(w, "Video is not available", http.StatusBadRequest)
			return
		}
	}

	// Check if the video is longer than 180 minutes
	if isValid, err := c.YVideoprofiler.CheckVideoDuration(ctx, url, 144000); err != nil || !isValid {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !isValid {
			http.Error(w, "Video is longer than 180 minutes", http.StatusBadRequest)
			return
		}
	}

	resultChan := make(chan convertMp3Result)

	downloadCtx, downloadCancel := context.WithTimeout(context.Background(), 30*time.Minute)
	downloadResults[downloadToken] = []byte{}
	downloadProgress[downloadToken] = &DownloadProgress{
		progress:   false,
		lastCheck:  time.Now(),
		CancelFunc: downloadCancel,
	}

	go func() {
		resultChan = c.ConvertMp3Async(downloadCtx, url, downloadToken)
		result := <-resultChan
		downloadResults[downloadToken] = result.Result
		if downloadProgressToken, ok := downloadProgress[downloadToken]; ok {
			downloadProgressToken.SetProgress(true)
		} else {
			log.Println("Download progress could not be found....")
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	response := struct {
		Token             string `json:"token"`
		HealthCheckPeriod int    `json:"health_check_period"`
	}{
		Token:             downloadToken,
		HealthCheckPeriod: 5000,
	}
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error generating download token", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// ProgressHandler to get the download progress
func (c *YoutubeConvertorController) ProgressHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if progress, ok := downloadProgress[token]; ok {
		currentProgress := progress.GetProgress()
		w.Header().Set("Content-Type", "text/plain")
		if !currentProgress {
			w.WriteHeader(http.StatusAccepted) // Set status code to 202 when the download is still in progress
		} else {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		http.Error(w, "Invalid token or download not found", http.StatusNotFound)
	}
}

// DownloadResultHandler to get the completed file
func (c *YoutubeConvertorController) DownloadResultHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	log.Println("Token from donwload", token)
	if result, ok := downloadResults[token]; ok {
		// Set the response header to indicate the file download
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("%s.mp3", time.Now().String()))
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(result)))
		w.Write(result)
	} else {
		http.Error(w, "Invalid token or download not found", http.StatusNotFound)
	}
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

func (c *YoutubeConvertorController) ConvertMp3Async(ctx context.Context, url string, token string) chan convertMp3Result {
	result := make(chan convertMp3Result, 1)

	go func() {
		// Call DownloadMp3
		mp3File, _, err := c.Mp3downloader.DownloadMp3(ctx, url)

		if err != nil {
			log.Println("Error in DownloadMp3:", err)
		}
		// Send the result to result channel
		res := convertMp3Result{
			Result: mp3File,
			Err:    err,
		}
		result <- res
		close(result) // Close the channel to signal that the operation has completed
	}()

	return c.waitForResult(ctx, result, token)
}

func (c *YoutubeConvertorController) waitForResult(ctx context.Context, result chan convertMp3Result, token string) chan convertMp3Result {
	for {
		select {
		case <-ctx.Done():
			return c.returnResult(result)
		case res := <-result:
			return c.createResultChannel(res)
		case <-time.After(1 * time.Second):
			if time.Since(downloadProgress[token].GetLastCheck()) > 6*time.Second {
				downloadProgress[token].CancelFunc() // Cancel the download if no progress check within the timeout duration
				return nil
			}
			time.Sleep(500 * time.Millisecond) // Add a sleep to prevent high CPU usage
		}
	}
}

func (c *YoutubeConvertorController) returnResult(result chan convertMp3Result) chan convertMp3Result {
	res, ok := <-result
	if ok {
		return c.createResultChannel(res)
	}
	return nil
}

func (c *YoutubeConvertorController) createResultChannel(res convertMp3Result) chan convertMp3Result {
	resChan := make(chan convertMp3Result, 1)
	resChan <- res
	close(resChan)
	return resChan
}
