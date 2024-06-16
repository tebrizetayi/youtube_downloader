package mp3downloader

import (
	"context"
	"strings"
	"youtube_download/internal/convertor"
	"youtube_download/internal/downloader"

	"go.uber.org/zap"
)

type Mp3downloader interface {
	DownloadMp3(ctx context.Context, url string) ([]byte, string, error)
}

type Client struct {
	Downloader downloader.Downloader
	Converter  convertor.Converter
	Logger     *zap.Logger
}

func NewMp3downloader(d downloader.Downloader, c convertor.Converter, logger *zap.Logger) Client {
	return Client{
		Downloader: d,
		Converter:  c,
		Logger:     logger,
	}
}

func (c *Client) DownloadMp3(ctx context.Context, url string) ([]byte, string, error) {
	fileNamemp4, err := c.Downloader.Download(ctx, url)
	if err != nil {
		return nil, "", err
	}

	filemp3 := strings.ReplaceAll(fileNamemp4, ".mp4", ".mp3")
	mp3Bytes, err := c.Converter.ConvertMp4ToMp3(ctx, fileNamemp4, filemp3)
	if err != nil {
		return nil, "", err
	}

	return mp3Bytes, filemp3, nil
}
