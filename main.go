package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"youtube_download/api"
	"youtube_download/youtube"
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
	youtubeClient := youtube.NewYoutubeClient()
	controller := api.NewYoutubeController(&youtubeClient)

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

	// =========================================================================
	// Shutdown
	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("main : Error starting server: %+v", err)

	case sig := <-shutdown:
		log.Printf("main : %v : Start shutdown..", sig)
	}
}

// ConvertIntToString : Convert int to string
func ConvertIntToString(i int) string {
	return strconv.Itoa(i)
}
