package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	"youtube_download/api"
	"youtube_download/internal/convertor"
	"youtube_download/internal/downloader"
	"youtube_download/internal/mp3downloader"
	"youtube_download/internal/youtubevideoprofiler"
)

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	port = ":" + port

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Services
	downloader := downloader.NewDownloader()
	convertor := convertor.NewConverter()
	mp3downloader := mp3downloader.NewMp3downloader(&downloader, &convertor)
	youtubevideoprofiler := youtubevideoprofiler.NewVideoProfiler()
	controller := api.NewYoutubeController(&mp3downloader, youtubevideoprofiler)

	// Start the HTTP service listening for requests.
	api := http.Server{
		Addr:           port,
		Handler:        api.NewAPI(controller),
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("main : API Listening %s", port)
		serverErrors <- api.ListenAndServe()
	}()

	root := "/go/src/app"
	duration := 10 * time.Minute
	ticker := time.NewTicker(duration) // Set up a ticker that ticks every 15 minutes
	defer ticker.Stop()                // Ensure the ticker is stopped to free resources

	for {
		select {
		case <-ticker.C: // Wait for the next tick
			fmt.Println("Performing scheduled file check and deletion...")
			filepath.Walk(root, deleteOldFiles)

		case err := <-serverErrors:
			log.Fatalf("main : Error starting server: %+v", err)

		case sig := <-shutdown:
			log.Printf("main : %v : Start shutdown..", sig)
		}
	}
}

// ConvertIntToString : Convert int to string
func ConvertIntToString(i int) string {
	return strconv.Itoa(i)
}

func deleteOldFiles(path string, fileInfo os.FileInfo, err error) error {
	duration := 10 * time.Minute
	if err != nil {
		log.Println(err) // print any error but continue
		return nil
	}
	log.Println("path:", path)

	// Check if the file is an mp3 or mp4
	if filepath.Ext(path) == ".mp3" || filepath.Ext(path) == ".mp4" {
		// Get the creation time of the file
		stat, err := os.Stat(path)
		if err != nil {
			log.Println(err)
			return nil
		}

		log.Println(stat.ModTime())
		log.Println(time.Since(stat.ModTime()))
		// Calculate time difference
		if time.Since(stat.ModTime()) > duration {
			// If the file is older than 15 minutes, delete it
			log.Println("Deleting:", path)
			err := os.Remove(path)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
