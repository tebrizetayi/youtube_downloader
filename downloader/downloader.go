package downloader

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/iawia002/lux/app"
	//"github.com/rylio/ytdl"
)

var (
	ErrExceedingDurationLimits = fmt.Errorf("video is exceeding the limits")
)

type Downloader interface {
	Download(ctx context.Context, url string) (string, error)
}

type Client struct {
}

func NewDownloader() Client {
	return Client{}
}

// Renaming the filename
func (c *Client) RenameVideoFileName(videoFileName string) string {
	fileName := videoFileName + fmt.Sprintf("%d.mp4", time.Now().UnixNano())
	re := regexp.MustCompile(`[/\\:*?"<>|\s()]`)
	fileName = strings.ToLower(re.ReplaceAllString(fileName, "_"))
	return fileName
}

func (c *Client) Download(ctx context.Context, url string) (string, error) {
	fileName := fmt.Sprintf("%d", time.Now().UnixNano())
	if err := app.New().Run([]string{"main", "--multi-thread", "-f", "140", "-O", fileName, url}); err != nil {
		fmt.Fprintf(
			color.Output,
			"Run %s failed: %s\n",
			color.CyanString("%s", app.Name), color.RedString("%v", err),
		)
		return "", err
	}

	return fileName, nil
}
