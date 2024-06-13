package mp3downloader

import (
	"context"
	"log"
	"strings"
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

	log.Println("BEGIN")
	fileNamemp4, err := c.Downloader.Download(ctx, url)
	if err != nil {
		return nil, "", err
	}

	log.Println(fileNamemp4)
	log.Println("Video is downloaded")

	filemp3 := strings.ReplaceAll(fileNamemp4, ".mp4", ".mp3")
	mp3Bytes, err := c.Converter.ConvertMp4ToMp3(ctx, fileNamemp4, filemp3)
	if err != nil {
		return nil, "", err
	}

	log.Println("ENDED")
	return mp3Bytes, filemp3, nil
}
