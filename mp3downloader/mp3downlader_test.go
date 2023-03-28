package mp3downloader

import (
	"os"
	"strings"
	"testing"
	"youtube_download/convertor"
	"youtube_download/downloader"
)

func TestMp3Downloader_Success(t *testing.T) {

	testTable := []struct {
		URL string
		err error
	}{
		{
			err: nil,
			URL: "https://www.youtube.com/watch?v=8aw6lLu-iBo",
		},
	}
	downloader := downloader.NewDownloader()
	convertor := convertor.NewConverter()
	mp3downloader := NewMp3downloader(&downloader, &convertor)
	for _, test := range testTable {
		_, mp3Filename, err := mp3downloader.DownloadMp3(test.URL)
		if test.err != err {
			t.Error("error during downloading youtube video to mp3", err)
			return
		}

		if err := os.Remove(mp3Filename); err != nil {
			t.Error("error during removing downloaded video", err)
			return
		}

		if err := os.Remove(strings.ReplaceAll(mp3Filename, ".mp3", ".mp4")); err != nil {
			t.Error("error during removing downloaded mp3", err)
			return
		}
	}

}
