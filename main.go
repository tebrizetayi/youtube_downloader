package main

import (
	"fmt"
	"net/http"
	"youtube_download/api"
	"youtube_download/youtube"
)

func main() {

	youtubeClient := youtube.NewYoutubeClient()
	youtubeController := api.NewYoutubeController(&youtubeClient)
	// Register the download handler function
	http.HandleFunc("/download", youtubeController.DownloadMp3)
	http.HandleFunc("/", youtubeController.ServeIndex)
	// Start the web server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Failed to start web server:", err)
	}
}
