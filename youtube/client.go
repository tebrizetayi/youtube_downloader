package youtube

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/iawia002/lux/app"
	"github.com/kkdai/youtube/v2"
	//"github.com/rylio/ytdl"
)

var (
	ErrExceedingDurationLimits = fmt.Errorf("video is exceeding the limits")
)

type YoutubeDownloader interface {
	DownloadYouTubeMP3(url string) ([]byte, error)
}

type Client struct {
}

func NewYoutubeClient() Client {
	return Client{}
}

func (c *Client) DownloadYouTubeMP3(url string) ([]byte, error) {

	filename, err := c.downloadLux(url)
	if err != nil {
		return nil, err
	}

	mp3Bytes, err := c.convertMp4ToMp3(filename)
	if err != nil {
		return nil, err
	}

	return mp3Bytes, nil
}

func (c *Client) download(videoID string) (string, error) {
	log.Println("downloading URL ", videoID)
	clientYoutube := youtube.Client{}

	video, err := clientYoutube.GetVideo(videoID)
	if err != nil {
		panic(err)
	}
	if video.Duration.Seconds() > 600.0 {
		return "", ErrExceedingDurationLimits
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	//testDownloader.DownloadComposite(ctx, "", video, "hd1080", "mp4")

	//formats[0].AudioQuality
	stream, _, err := clientYoutube.GetStream(video, &formats.Type("audio/mp4")[0])
	if err != nil {
		return "", err
	}
	fileName := renameVideoFileName(video.Title)

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return "", err
	}
	log.Println("downloaded URL ", videoID)
	return fileName, nil
}

func (c *Client) convertMp4ToMp3(fileName string) ([]byte, error) {
	// Convert video to MP3

	log.Println("Converting file to mp3", fileName)
	mp3File := fmt.Sprintf("%s.mp3", fileName)
	//ffmpeg -i input_video.mp4 -vn -acodec libmp3lame -b:a 128k -ar 44100 -ac 2 -threads 4 -preset ultrafast output_audio.mp3

	//cmd := exec.Command("ffmpeg", "-hwaccel", "auto", "-i", fileName, "-vn", "-acodec", "libmp3lame", "-b:a", "96k", "-ar", "44100", "-ac", "2", "-s", "640x360", "-threads", "8", "-preset", "ultrafast", mp3File)

	//cmd := exec.Command("ffmpeg", "-i", fileName, "-vn", "-acodec", "libmp3lame",
	//	"-b:a", "128k", "-ar", "44100", "-ac", "2", "-threads", "4", "-preset ultrafast", mp3File)
	//cmd := exec.Command("ffmpeg", "-i", fileName, mp3File)

	cmd := exec.Command("ffmpeg", "-i", fileName+".mp4", "-vn", "-acodec", "libmp3lame", "-ac", "2",
		"-ab", "256k", "-ar", "44100", mp3File)
	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	mp3Bytes, err := os.ReadFile(fmt.Sprintf("%s.mp3", fileName))
	if err != nil {
		return nil, err
	}

	log.Println("Converted file to mp3", fileName)
	return mp3Bytes, nil

}

// Renaming the filename
func renameVideoFileName(videoFileName string) string {
	fileName := videoFileName + fmt.Sprintf("%d.mp4", time.Now().UnixNano())
	re := regexp.MustCompile(`[/\\:*?"<>|\s()]`)
	fileName = strings.ToLower(re.ReplaceAllString(fileName, "_"))
	return fileName
}

func (c *Client) downloadLux(url string) (string, error) {
	fileName := fmt.Sprintf("%d", time.Now().UnixNano())

	if err := app.New().Run([]string{"main", "--multi-thread", "-f", "140", "-O", fileName, url}); err != nil {
		fmt.Fprintf(
			color.Output,
			"Run %s failed: %s\n",
			color.CyanString("%s", app.Name), color.RedString("%v", err),
		)
		return "", err
	}

	return fileName, nil
}
