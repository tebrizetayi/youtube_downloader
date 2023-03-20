package youtube

import (
	"testing"
)

func TestYoutubeMp3_Success(t *testing.T) {
	youtubeClient := NewYoutubeClient()
	_, err := youtubeClient.DownloadYouTubeMP3("https://www.youtube.com/watch?v=8aw6lLu-iBo")
	if err != nil {
		t.Error("error during converting youtube video to mp3", err)
		return
	}
}
