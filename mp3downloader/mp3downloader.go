package mp3downloader

import (
	"youtube_download/convertor"
	"youtube_download/downloader"
)

type Mp3downloader interface {
	DownloadMp3(url string) ([]byte, string, error)
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

func (c *Client) DownloadMp3(url string) ([]byte, string, error) {
	mp4FileName, err := c.Downloader.Download(url)
	if err != nil {
		return nil, "", err
	}

	outputFilename := mp4FileName
	mp3Bytes, err := c.Converter.ConvertMp4ToMp3(mp4FileName, outputFilename)
	if err != nil {
		return nil, "", err
	}

	return mp3Bytes, mp4FileName + ".mp3", nil
}
