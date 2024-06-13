package api

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"net/http"
	"net/url"
	"time"
	"youtube_download/internal/mp3downloader"
	"youtube_download/internal/youtubevideoprofiler"
)

type YoutubeConvertorController struct {
	mp3downloader.Mp3downloader
	YVideoprofiler  youtubevideoprofiler.ProfilerClient
	downloadManager *DownloadManager
}

func NewYoutubeController(mp3Downloader mp3downloader.Mp3downloader,
	youtubevideoprofiler youtubevideoprofiler.ProfilerClient,
) YoutubeConvertorController {
	return YoutubeConvertorController{
		mp3Downloader,
		youtubevideoprofiler,
		NewDownloadManager(),
	}
}

// DownloadMp3
func (c *YoutubeConvertorController) DownloadMp3(w http.ResponseWriter, r *http.Request) {
	ip, err := getIP(r)
	if err != nil {
		http.Error(w, "Error getting IP", http.StatusBadRequest)
		return
	}
	log.Printf("Download mp3 request received from Ip:%s\n", ip)

	referer := r.Referer()
	log.Printf("Referer: %s", referer)
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
	downloadToken := c.GenerateHash(url)

	if v, ok := c.downloadManager.progress[downloadToken]; !ok || !v.isDownloadCompleted {
		c.downloadManager.progress[downloadToken] = &DownloadProgress{
			isDownloadCompleted: true,
			lastCheck:           time.Now(),
		}
	} else {
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
		return
	}

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
	if isValid, err := c.YVideoprofiler.CheckVideoDuration(ctx, url, 900); err != nil || !isValid {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !isValid {
			http.Error(w, "Video is longer than 15 minutes", http.StatusBadRequest)
			return
		}
	}

	resultChan := make(chan convertMp3Result)

	downloadCtx, downloadCancel := context.WithTimeout(context.Background(), 30*time.Minute)
	c.downloadManager.progress[downloadToken] = &DownloadProgress{
		isDownloadCompleted: false,
		lastCheck:           time.Now(),
		CancelFunc:          downloadCancel,
		result:              []byte{},
	}

	go func() {
		resultChan = c.ConvertMp3Async(downloadCtx, url, downloadToken)
		result := <-resultChan
		c.downloadManager.SetResult(downloadToken, result.Result)
		if downloadProgressToken, ok := c.downloadManager.progress[downloadToken]; ok {
			downloadProgressToken.isDownloadCompleted = true
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
	if isDownloadCompleted, err := c.downloadManager.GetProgress(token); err == nil {
		if !isDownloadCompleted {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusAccepted) // Set status code to 202 when the download is still in progress
		} else {
			w.Header().Set("Content-Type", "audio/mpeg")
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
	if result, err := c.downloadManager.GetResult(token); err == nil {
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
			if lastCheck, err := c.downloadManager.GetLastCheck(token); err == nil {
				if time.Since(lastCheck) > 6*time.Second {
					err := c.downloadManager.CancelDownload(token)
					if err != nil {
						log.Println("Error in cancel download", err)
					}
					return nil
				}
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

func (c *YoutubeConvertorController) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// Open the file to be downloaded
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Extract watch?v= parameter from the URL
	videoParam := r.URL.Query().Get("v")
	if videoParam != "" {
		r.PostForm = url.Values{}
		r.PostForm.Set("url", fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoParam))
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	url := r.FormValue("url")
	log.Println("Downloading:", url)

	// Generate a unique download token
	downloadToken := url
	c.downloadManager.progress[downloadToken] = &DownloadProgress{
		isDownloadCompleted: true,
		lastCheck:           time.Now(),
	}

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
	if isValid, err := c.YVideoprofiler.CheckVideoDuration(ctx, url, 900); err != nil || !isValid {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !isValid {
			http.Error(w, "Video is longer than 15 minutes", http.StatusBadRequest)
			return
		}
	}

	resultChan := make(chan convertMp3Result)

	downloadCtx, downloadCancel := context.WithTimeout(context.Background(), 30*time.Minute)
	c.downloadManager.progress[downloadToken] = &DownloadProgress{
		isDownloadCompleted: false,
		lastCheck:           time.Now(),
		CancelFunc:          downloadCancel,
		result:              []byte{},
	}

	go func() {
		resultChan = c.ConvertMp3Async(downloadCtx, url, downloadToken)
		result := <-resultChan
		c.downloadManager.SetResult(downloadToken, result.Result)
		if downloadProgressToken, ok := c.downloadManager.progress[downloadToken]; ok {
			downloadProgressToken.isDownloadCompleted = true
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

// DownloadMp3
func (c *YoutubeConvertorController) WatchHandler(w http.ResponseWriter, r *http.Request) {
	// Open the file to be downloaded

	// Extract watch?v= parameter from the URL
	videoParam := r.URL.Query().Get("v")
	log.Println(videoParam)
	if videoParam != "" {

		r.PostForm = url.Values{}
		r.PostForm.Set("url", videoParam)
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	url := r.FormValue("url")

	if url == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Println("Downloading:", url)

	// Replace JSON response with HTTP redirection
	http.Redirect(w, r, fmt.Sprintf("?v=%s", url), http.StatusSeeOther)
}

func (c YoutubeConvertorController) GenerateHash(str string) string {
	hashSum := md5.Sum([]byte(str))
	return hex.EncodeToString(hashSum[:])
}

func getIP(r *http.Request) (string, error) {
	//Get IP from the X-REAL-IP header
	ip := r.Header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	//Get IP from X-FORWARDED-FOR header
	ips := r.Header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	//Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", err
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}
	return "", fmt.Errorf("no valid ip found")
}
