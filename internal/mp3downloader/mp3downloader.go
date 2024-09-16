package mp3downloader

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
	"youtube_download/internal/convertor"

	"go.uber.org/zap"
)

func (c *Client) fileExists(filePath string) bool {
	// os.Stat returns the file info or an error if it doesn't exist
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist
			return false
		}
		// Another error occurred, handle it
		c.Logger.Info("Error checking file:", zap.Error(err))
		return false
	}
	// File exists
	return true
}

type Mp3downloader interface {
	DownloadMp3(ctx context.Context, url string) ([]byte, string, error)
}

type Client struct {
	Converter convertor.Converter
	Logger    *zap.Logger
}

func NewMp3downloader(c convertor.Converter, logger *zap.Logger) Client {
	return Client{
		Converter: c,
		Logger:    logger,
	}
}
func (c *Client) DownloadMp3(ctx context.Context, url string) ([]byte, string, error) {
	//cookiesPath := "~/cookies-youtube-com.txt"
	cookiesFile := "/go/src/app/cookies-youtube-com.txt"

	if c.fileExists(cookiesFile) {
		c.Logger.Info("Cookies file exists!")
	} else {
		c.Logger.Info("Cookies file does not exist!")
	}
	fileName := fmt.Sprintf("%d", time.Now().UnixNano())

	url, err := c.ExtractVideoID(url)
	if err != nil {
		return nil, "", err
	}

	//yt-dlp -o "myvideo.mp4" -f "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best" https://www.youtube.com/watch?v=dQw4w9WgXcQ

	// Correctly separate the '-f' and its argument without single quotes around the format specifier
	//cmd := exec.CommandContext(ctx, "youtube-dl", "-f", "best[ext=mp4]", "-o", fileName, url)
	//cmd := exec.CommandContext(ctx, "yt-dlp", "-x", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", fileName, url)
	cmd := exec.CommandContext(ctx, "yt-dlp", "-vU", "-x", "--extractor-args", "youtube:no-video-proxy", "--cookies", cookiesFile, "--audio-format", "mp3", "-o", fileName, url)

	//yt-dlp -x --audio-format mp3 -o "random.mp3"  https://www.youtube.com/watch?v=UD3t3nY9xJ8

	c.Logger.Info("executing command", zap.Any("cmd", cmd.Args))

	// Start the command
	err = cmd.Start()
	if err != nil {
		return nil, "", fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for command to complete or context cancellation
	select {
	case <-ctx.Done():
		// If context is done, attempt to kill the process
		if killErr := cmd.Process.Kill(); killErr != nil {
			return nil, "", fmt.Errorf("failed to kill process: %w", killErr)
		}
		return nil, "", ctx.Err()

	default:
		// Wait for the command to finish executing
		err = cmd.Wait()
		if err != nil {
			return nil, "", fmt.Errorf("error waiting for command to finish: %w", err)
		}
	}

	mp3Bytes, err := os.ReadFile(fileName + ".mp3")
	if err != nil {
		return nil, "", err
	}

	return mp3Bytes, fileName + ".mp3", nil
}

func (c *Client) ExtractVideoID(inputURL string) (string, error) {
	// Parse the input URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		return "", err
	}

	// Check the host and path to determine if it's a YouTube URL
	if parsedURL.Host != "www.youtube.com" && parsedURL.Host != "youtube.com" {
		return "", fmt.Errorf("not a YouTube URL")
	}

	var videoID string

	// Check if it's a short URL
	if strings.HasPrefix(parsedURL.Path, "/shorts/") {
		videoID = strings.TrimPrefix(parsedURL.Path, "/shorts/")
	} else {
		// Extract the video ID from query parameters for other YouTube URLs
		queryParams := parsedURL.Query()
		videoID = queryParams.Get("v")
		if videoID == "" {
			return "", fmt.Errorf("no video ID found in URL")
		}
	}

	// Construct the standard YouTube link
	standardURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s&", videoID)
	return standardURL, nil
}
