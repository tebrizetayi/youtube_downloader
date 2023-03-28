package youtubevideoprofiler

import (
	"context"
	"testing"
)

func TestYVideoProfiler_Success(t *testing.T) {
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
	yvideoProfiler := NewYoutubevideoprofiler()
	for _, test := range testTable {
		_, err := yvideoProfiler.Info(ctx, test.URL)
		if test.err != err {
			t.Error("error during getting video profile", err)
			return
		}
	}
}

func TestYVideoProfiler_Duration(t *testing.T) {
	ctx := context.Background()
	testTable := []struct {
		URL              string
		duration         float64
		isWithinDuration bool
		err              error
	}{
		{
			URL:              "https://www.youtube.com/watch?v=8aw6lLu-iBo",
			duration:         10,
			isWithinDuration: true,
			err:              nil,
		},
		{
			URL:              "https://www.youtube.com/watch?v=fYU-cz9j61g",
			duration:         10,
			isWithinDuration: false,
			err:              nil,
		},
	}
	yvideoProfiler := NewYoutubevideoprofiler()
	for _, test := range testTable {
		isWithinDuration, err := yvideoProfiler.CheckDuration(ctx, test.URL, test.duration)
		if test.err != err || test.isWithinDuration != isWithinDuration {
			t.Error("error cheking video profile", err)
			return
		}

	}
}
