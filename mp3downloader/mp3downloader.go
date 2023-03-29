package mp3downloader

import (
	"context"
	"youtube_download/convertor"
	"youtube_download/downloader"
)

type Mp3downloader interface {
	DownloadMp3(ctx context.Context, url string) ([]byte, string, error)
	//ConvertorMp3(ctx context.Context, url string) (string, error)
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
	mp4FileName, err := c.Downloader.Download(ctx, url)
	if err != nil {
		return nil, "", err
	}

	outputFilename := mp4FileName
	mp3Bytes, err := c.Converter.ConvertMp4ToMp3(ctx, mp4FileName, outputFilename)
	if err != nil {
		return nil, "", err
	}
	return mp3Bytes, mp4FileName + ".mp3", nil
}

/*
func (c *Client) ConvertorMp3(ctx context.Context, url string) (string, error) {
	mp4FileName, err := c.Downloader.Download(ctx, url)
	if err != nil {
		return "", err
	}

	outputFilename := mp4FileName
	_, err = c.Converter.ConvertMp4ToMp3(ctx, mp4FileName, outputFilename)
	if err != nil {
		return "", err
	}

	return mp4FileName + ".mp3", nil
}*/
