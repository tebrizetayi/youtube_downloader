package downloader

import (
	"context"
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestDownloader_Success(t *testing.T) {
	ctx := context.Background()
	testTable := []struct {
		URL string
		err error
	}{
		{
			err: nil,
			URL: "https://www.youtube.com/watch?v=RtBy-7aGiMo",
		},
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	downloader := NewDownloader(logger)
	for _, test := range testTable {
		fileName, err := downloader.Download(ctx, test.URL)
		if test.err != err {
			t.Error("error during downloading video", err)
			return
		}

		if err := os.Remove(fileName); err != nil {
			t.Error("error during removing downloaded video", err)
			return
		}

	}

}

/*
func BenchmarkYoutubeMp3(b *testing.B) {
	b.Skip()
	for i := 0; i < b.N; i++ {
		youtubeClient := NewYoutubeClient()
		_, err := youtubeClient.DownloadYouTubeMP3("https://www.youtube.com/watch?v=QTUDVuNuKss")
		if err != nil {
			b.Error("error during converting youtube video to mp3", err)
			return
		}
	}
}

func BenchmarkYoutubeDownload(b *testing.B) {
	b.Skip()
	for i := 0; i < b.N; i++ {
		youtubeClient := NewYoutubeClient()
		_, err := youtubeClient.downloadLux("https://www.youtube.com/watch?v=NVh_wS7ECsM")
		if err != nil {
			b.Error("error during downloading youtube video", err)
			return
		}
	}
}

func TestYoutubeDownload(t *testing.T) {
	t.Skip()
	youtubeClient := NewYoutubeClient()
	_, err := youtubeClient.downloadLux("https://www.youtube.com/watch?v=NVh_wS7ECsM")
	if err != nil {
		t.Error("error during downloading youtube video", err)
		return
	}

}

*/

/*
func BenchmarkYoutubeDownload(b *testing.B) {
	b.Skip()
	ctx := context.Background()
	downloader := NewDownloader()
	for i := 0; i < b.N; i++ {
		fileName, err := downloader.Download(ctx, "https://www.youtube.com/watch?v=fYU-cz9j61g")
		if err != nil {
			b.Error("error during downloading video", err)
			return
		}

		if err := os.Remove(fileName + ".mp4"); err != nil {
			b.Error("error during removing downloaded video", err)
			return
		}
	}
}
*/
