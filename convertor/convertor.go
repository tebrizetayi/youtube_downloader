package convertor

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Converter interface {
	ConvertMp4ToMp3(ctx context.Context, fileName string, outputFileName string) ([]byte, error)
}
type Client struct {
}

func NewConverter() Client {
	return Client{}
}

func (c *Client) ConvertMp4ToMp3(ctx context.Context, fileName string, outputFilename string) ([]byte, error) {
	// Convert video to MP3

	log.Println("Converting file to mp3", fileName)

	mp3File := fmt.Sprintf("%s.mp3", outputFilename)
	cmd := exec.Command("ffmpeg", "-i", fileName+".mp4", "-vn", "-acodec", "libmp3lame", "-ac", "2",
		"-ab", "256k", "-ar", "44100", mp3File)
	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	mp3Bytes, err := os.ReadFile(mp3File)
	if err != nil {
		return nil, err
	}

	log.Println("Converted file to mp3", mp3File)
	return mp3Bytes, nil

}
