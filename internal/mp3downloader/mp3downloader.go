package mp3downloader

import (
	"context"
	"youtube_download/internal/convertor"
	"youtube_download/internal/downloader"
)

type Mp3downloader interface {
	DownloadMp3(ctx context.Context, url string) ([]byte, string, error)
}

type Client struct {
	Downloader downloader.Downloader
	Converter  convertor.Converter
}

func NewMp3downloader(d downloader.Downloader, c convertor.Converter) Client {
	return Client{
		Downloader: d,
		Converter:  c,
	}
}

func (c *Client) DownloadMp3(ctx context.Context, url string) ([]byte, string, error) {

	fileName, err := c.Downloader.Download(ctx, url)
	if err != nil {
		return nil, "", err
	}

	mp3Bytes, err := c.Converter.ConvertMp4ToMp3(ctx, fileName, fileName)
	if err != nil {
		return nil, "", err
	}

	return mp3Bytes, fileName + ".mp3", nil
}
