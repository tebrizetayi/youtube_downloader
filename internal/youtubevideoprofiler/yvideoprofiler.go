package youtubevideoprofiler

import (
	"context"
	"errors"

	"github.com/kkdai/youtube/v2"
	"go.uber.org/zap"
)

var (
	ErrVideoNotFound = errors.New("video is not found")
)

type VideoProfiler interface {
	GetVideoInfo(ctx context.Context, videoID string) (*youtube.Video, error)
	CheckVideoDuration(ctx context.Context, videoID string, maxDuration float64) (bool, error)
	IsVideoAvailable(ctx context.Context, videoID string) (bool, error)
}

type ProfilerClient struct {
	logger *zap.Logger
}

func NewVideoProfiler(logger *zap.Logger) ProfilerClient {
	return ProfilerClient{
		logger: logger,
	}
}

func (c *ProfilerClient) GetVideoInfo(ctx context.Context, videoID string) (*youtube.Video, error) {
	youtubeClient := youtube.Client{}

	video, err := youtubeClient.GetVideo(videoID)
	if err != nil {
		return nil, err
	}

	return video, nil
}

func (c *ProfilerClient) CheckVideoDuration(ctx context.Context, videoID string, maxDuration float64) (bool, error) {
	video, err := c.GetVideoInfo(ctx, videoID)
	if err != nil {
		return false, err
	}
	if video.Duration.Seconds() > maxDuration {
		return false, nil
	}

	return true, nil
}

func (c *ProfilerClient) IsVideoAvailable(ctx context.Context, videoID string) (bool, error) {
	_, err := c.GetVideoInfo(ctx, videoID)
	if err != nil {
		c.logger.Error("error ", zap.Error(err))
		//return false, ErrVideoNotFound
	}
	return true, nil
}
