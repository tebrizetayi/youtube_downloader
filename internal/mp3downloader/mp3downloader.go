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
	cookiesFile := "/go/src/app/cookies-youtube-com.txt"

	// Check if the cookies file exists
	if c.fileExists(cookiesFile) {
		c.Logger.Info("Cookies file exists!")
	} else {
		c.Logger.Info("Cookies file does not exist!")
	}

	// Generate a unique filename based on the current time
	fileName := fmt.Sprintf("%d", time.Now().UnixNano())

	// Extract the video ID from the URL
	url, err := c.ExtractVideoID(url)
	if err != nil {
		return nil, "", err
	}

	// Prepare the yt-dlp command to download the audio as mp3
	cmd := exec.CommandContext(ctx, "yt-dlp", "-vU", "-v", "-x", "--audio-format", "mp3", "-o", fileName+".mp3", url)

	c.Logger.Info("executing command", zap.Any("cmd", cmd.Args))

	// Capture the output (stdout and stderr) from the command execution
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Log the command output (stdout/stderr) in case of an error
		c.Logger.Error("failed to execute command", zap.Error(err), zap.String("output", string(output)))
		return nil, "", fmt.Errorf("failed to execute command: %w, output: %s", err, string(output))
	}

	// Check if the output file was created successfully
	outputFile := fileName + ".mp3"
	if !c.fileExists(outputFile) {
		return nil, "", fmt.Errorf("output file %s was not created", outputFile)
	}

	// Read the output MP3 file
	mp3Bytes, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, "", err
	}

	// Clean up the file after reading it
	err = os.Remove(outputFile)
	if err != nil {
		c.Logger.Warn("failed to remove temporary file", zap.String("file", outputFile), zap.Error(err))
	}

	return mp3Bytes, outputFile, nil
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
