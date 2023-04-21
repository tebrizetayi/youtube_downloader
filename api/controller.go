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

	"github.com/google/uuid"
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
	progress float64
}

// Store the download progress and result in a map
var downloadProgress = make(map[string]*DownloadProgress)
var downloadResults = make(map[string][]byte)

func (dp *DownloadProgress) SetProgress(progress float64) {
	log.Println("Setting for 100 % percent")
	dp.Lock()
	defer dp.Unlock()
	log.Println("Setting for 100 % percent.....")
	dp.progress = progress
}

func (dp *DownloadProgress) GetProgress() float64 {
	dp.Lock()
	defer dp.Unlock()
	return dp.progress
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
	downloadToken := uuid.New().String()
	progress := &DownloadProgress{}
	downloadProgress[downloadToken] = progress

	// Check if the video is exists and public
	IsAvailable, err := c.YVideoprofiler.IsVideoAvailable(ctx, url)
	if !IsAvailable {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Check if the video is longer than 10 minutes
	isValid, _ := c.YVideoprofiler.CheckVideoDuration(ctx, url, 1200)
	if !isValid {
		http.Error(w, "Video is longer than 10 minutes", http.StatusBadRequest)
		return
	}

	resultChan := make(chan convertMp3Result)

	downloadCtx, downloadCancel := context.WithTimeout(context.Background(), 30*time.Minute)
	_ = downloadCancel
	downloadResults[downloadToken] = []byte{}
	go func() {
		log.Println("Update the status")
		resultChan = c.ConvertMp3Async(downloadCtx, url, downloadToken)
		log.Println("Update the status 2")
		result := <-resultChan
		log.Println("Update the status  1")
		downloadResults[downloadToken] = result.Result
		log.Println("Update the status 6")

		if downloadProgressToken, ok := downloadProgress[downloadToken]; ok {
			log.Println("Download is done....")
			downloadProgressToken.SetProgress(100)
			log.Println("Download is done.... and progress is set to 100")
		} else {
			log.Println("Download progress could not be found....")
		}
	}()

	/*var result convertMp3Result
	select {
	case result = <-resultChan:
		if result.Err != nil {
			log.Println("Connection is errrorrr")
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	case <-ctx.Done():
		log.Println("Connection is closed")
		return
	}*/
	// Respond to the user immediately with the download token
	//w.WriteHeader(http.StatusAccepted)
	//w.Write([]byte(fmt.Sprintf("File download started, use token '%s' to check progress and download the file.", downloadToken)))

	// Set the response header to indicate the file download
	/*w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Disposition", "attachment; filename="+fmt.Sprintf("%s.mp3", time.Now().String()))
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(result.Result)))
	log.Println("Downloaded:", url)
	w.Write(result.Result)*/

	w.WriteHeader(http.StatusAccepted)
	response := struct {
		Token string `json:"token"`
	}{
		Token: downloadToken,
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
		if currentProgress < 100 {
			w.WriteHeader(http.StatusAccepted) // Set status code to 202 when the download is still in progress
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte(fmt.Sprintf("Current download progress: %.2f%%", currentProgress)))
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
		} else {
			log.Println("DownloadMp3 completed successfully")
		}

		log.Println("MP3 file is ready")
		// Send the result to result channel
		res := convertMp3Result{
			Result: mp3File,
			Err:    err,
		}
		result <- res
		close(result) // Close the channel to signal that the operation has completed
	}()

	return result
}
