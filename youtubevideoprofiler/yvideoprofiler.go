package youtubevideoprofiler

import (
	"github.com/kkdai/youtube/v2"
)

type YVideoprofiler interface {
	Info(videoID string) (*youtube.Video, error)
	CheckDuration(videoID string, timeConstraint float64) (bool, error)
}

type Client struct {
}

func NewYoutubevideoprofiler() Client {
	return Client{}
}

func (c *Client) Info(videoID string) (*youtube.Video, error) {
	clientYoutube := youtube.Client{}

	video, err := clientYoutube.GetVideo(videoID)
	if err != nil {
		return nil, err
	}

	return video, nil
}

func (c *Client) CheckDuration(videoID string, timeConstraint float64) (bool, error) {
	video, err := c.Info(videoID)
	if err != nil {
		return false, err
	}

	if video.Duration.Seconds() > timeConstraint {
		return false, nil
	}

	return true, nil
}
