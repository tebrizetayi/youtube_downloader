package youtubevideoprofiler

import (
	"context"
	"errors"
	"log"

	"github.com/kkdai/youtube/v2"
)

var (
	ErrVideoNotFound = errors.New("video not found")
)

type YVideoprofiler interface {
	Info(ctx context.Context, videoID string) (*youtube.Video, error)
	CheckDuration(ctx context.Context, videoID string, timeConstraint float64) (bool, error)
	IsAvailable(ctx context.Context, videoID string) (bool, error)
}

type Client struct {
}

func NewYoutubevideoprofiler() Client {
	return Client{}
}

func (c *Client) Info(ctx context.Context, videoID string) (*youtube.Video, error) {
	clientYoutube := youtube.Client{}

	video, err := clientYoutube.GetVideo(videoID)
	if err != nil {
		return nil, err
	}

	return video, nil
}

func (c *Client) CheckDuration(ctx context.Context, videoID string, timeConstraint float64) (bool, error) {
	video, err := c.Info(ctx, videoID)
	if err != nil {
		return false, err
	}

	if video.Duration.Seconds() > timeConstraint {
		return false, nil
	}

	return true, nil
}

func (c *Client) IsAvailable(ctx context.Context, videoID string) (bool, error) {
	_, err := c.Info(ctx, videoID)
	if err != nil {
		log.Println(err)
		return false, ErrVideoNotFound
	}
	return true, nil
}
