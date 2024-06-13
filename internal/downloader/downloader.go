package downloader

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"
	//"github.com/rylio/ytdl"
)

var (
	ErrExceedingDurationLimits = fmt.Errorf("")
)

type Downloader interface {
	Download(ctx context.Context, url string) (string, error)
}

type Client struct {
}

func NewDownloader() Client {
	return Client{}
}

func (c *Client) Download(ctx context.Context, url string) (string, error) {
	fileName := fmt.Sprintf("%d.mp4", time.Now().UnixNano())

	//yt-dlp -o "myvideo.mp4" -f "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best" https://www.youtube.com/watch?v=dQw4w9WgXcQ

	// Correctly separate the '-f' and its argument without single quotes around the format specifier
	//cmd := exec.CommandContext(ctx, "youtube-dl", "-f", "best[ext=mp4]", "-o", fileName, url)
	cmd := exec.CommandContext(ctx, "yt-dlp", "-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", fileName, url)
	log.Println("Executing command:", cmd.Args) // Logging the command arguments for debugging

	// Start the command
	err := cmd.Start()
	if err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for command to complete or context cancellation
	select {
	case <-ctx.Done():
		// If context is done, attempt to kill the process
		if killErr := cmd.Process.Kill(); killErr != nil {
			log.Println("Error killing process:", killErr)
			return "", fmt.Errorf("failed to kill process: %w", killErr)
		}
		return "", ctx.Err()

	default:
		// Wait for the command to finish executing
		err = cmd.Wait()
		if err != nil {
			return "", fmt.Errorf("error waiting for command to finish: %w", err)
		}
	}

	return fileName, nil
}
