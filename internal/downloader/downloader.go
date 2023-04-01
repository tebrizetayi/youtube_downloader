package downloader

import (
	"context"
	"fmt"
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
	fileName := fmt.Sprintf("%d", time.Now().UnixNano())

	cmd := exec.CommandContext(ctx, "lux", "--multi-thread", "-f", "140", "-O", fileName, url)
	err := cmd.Start()
	if err != nil {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	return fileName, nil
}
