package convertor

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDownloader_Success(t *testing.T) {
	ctx := context.Background()
	convertor := NewConverter()

	fileName := fmt.Sprintf("%d", time.Now().Unix())
	_, err := convertor.ConvertMp4ToMp3(ctx, "test", fileName)
	if err != nil {
		t.Error("error during converting video to mp3", err)
		return
	}

	if err := os.Remove(fileName + ".mp3"); err != nil {
		t.Errorf("error during removing mp3 filename: %s err:%s", fileName, err)
		return
	}

}
