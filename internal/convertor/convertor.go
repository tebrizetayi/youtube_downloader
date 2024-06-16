package convertor

import (
	"context"
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
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", fileName, "-vn", "-acodec", "libmp3lame", "-ac", "2",
		"-ab", "256k", "-ar", "44100", outputFilename)

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	mp3Bytes, err := os.ReadFile(outputFilename)
	if err != nil {
		return nil, err
	}
	return mp3Bytes, nil
}
