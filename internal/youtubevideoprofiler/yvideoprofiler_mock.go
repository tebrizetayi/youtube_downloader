package youtubevideoprofiler

import (
	"context"

	"github.com/kkdai/youtube/v2"
	"go.uber.org/zap"
)

type ProfilerClient_Mock struct {
	logger *zap.Logger
}

func NewVideoProfilerMock(logger *zap.Logger) ProfilerClient_Mock {
	return ProfilerClient_Mock{
		logger: logger,
	}
}

func (p ProfilerClient_Mock) GetVideoInfo(ctx context.Context, videoID string) (*youtube.Video, error) {
	// Implement the logic to profile a video using the videoID
	// This can include interacting with external services, processing data, etc.

	// For now, let's return a simple message as a placeholder
	return nil, nil
}

func (p ProfilerClient_Mock) CheckVideoDuration(ctx context.Context, videoID string, maxDuration float64) (bool, error) {
	return true, nil
}
func (p ProfilerClient_Mock) IsVideoAvailable(ctx context.Context, videoID string) (bool, error) {
	return true, nil
}
