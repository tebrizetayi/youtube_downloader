package mp3downloader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
	"youtube_download/internal/convertor"

	"go.uber.org/zap"
)

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
	fileName := fmt.Sprintf("%d", time.Now().UnixNano())

	//yt-dlp -o "myvideo.mp4" -f "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best" https://www.youtube.com/watch?v=dQw4w9WgXcQ

	// Correctly separate the '-f' and its argument without single quotes around the format specifier
	//cmd := exec.CommandContext(ctx, "youtube-dl", "-f", "best[ext=mp4]", "-o", fileName, url)
	//cmd := exec.CommandContext(ctx, "yt-dlp", "-x", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", "-o", fileName, url)
	cmd := exec.CommandContext(ctx, "yt-dlp", "-x", "--audio-format", "mp3", "-o", fileName, url)

	//yt-dlp -x --audio-format mp3 -o "random.mp3"  https://www.youtube.com/watch?v=UD3t3nY9xJ8

	c.Logger.Info("executing command", zap.Any("cmd", cmd.Args))

	// Start the command
	err := cmd.Start()
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
