package youtube

import (
	"testing"
)

func TestYoutubeMp3_Success(t *testing.T) {

	testTable := []struct {
		URL string
		err error
	}{
		{
			err: nil,
			URL: "https://www.youtube.com/watch?v=8aw6lLu-iBo",
		},
	}
	youtubeClient := NewYoutubeClient()
	for _, test := range testTable {
		_, err := youtubeClient.DownloadYouTubeMP3(test.URL)
		if test.err != err {
			t.Error("error during converting youtube video to mp3", err)
			return
		}
	}

}
func BenchmarkYoutubeMp3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		youtubeClient := NewYoutubeClient()
		_, err := youtubeClient.DownloadYouTubeMP3("https://www.youtube.com/watch?v=8aw6lLu-iBo")
		if err != nil {
			b.Error("error during converting youtube video to mp3", err)
			return
		}
	}
}
