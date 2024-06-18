package mp3downloader

import (
	"context"
	"os"
	"strings"
	"testing"
	"youtube_download/internal/convertor"

	"go.uber.org/zap"
)

func TestMp3Downloader_Success(t *testing.T) {

	ctx := context.Background()
	testTable := []struct {
		URL string
		err error
	}{
		{
			err: nil,
			URL: "https://www.youtube.com/watch?v=8aw6lLu-iBo",
		},
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	convertor := convertor.NewConverter()

	mp3downloader := NewMp3downloader(&convertor, logger)
	for _, test := range testTable {
		_, mp3Filename, err := mp3downloader.DownloadMp3(ctx, test.URL)
		if test.err != err {
			t.Error("error during downloading youtube video to mp3", err)
			return
		}

		if err := os.Remove(mp3Filename); err != nil {
			t.Error("error during removing downloaded video", err)
			return
		}
		//correct it.
		if err := os.Remove(strings.ReplaceAll(mp3Filename, ".mp3", ".mp4")); err != nil {
			t.Error("error during removing downloaded mp3", err)
			return
		}
	}
}
