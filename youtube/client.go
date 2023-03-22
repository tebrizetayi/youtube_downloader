package youtube

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

type YoutubeDownloader interface {
	DownloadYouTubeMP3(url string) ([]byte, error)
}

type Client struct {
}

func NewYoutubeClient() Client {
	return Client{}
}

func (c *Client) extractVideoID(url string) (string, error) {
	// Regular expression to match the video ID in the YouTube URL
	re := regexp.MustCompile(`(?:youtube(?:-nocookie)?\.com/(?:[^/\n\s]+/\S+/|(?:v|vi|e(?:mbed)?)/|\S*?[?&]v=)|youtu\.be/)([a-zA-Z0-9_-]{11})`)

	matches := re.FindStringSubmatch(url)
	if len(matches) == 0 {
		return "", fmt.Errorf("unable to extract video ID from URL")
	}

	return matches[1], nil
}

func (c *Client) DownloadYouTubeMP3(url string) ([]byte, error) {

	videoID, err := c.extractVideoID(url)
	if err != nil {
		return nil, err
	}
	filename, err := c.download(videoID)
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
	clientYoutube := youtube.Client{}

	video, err := clientYoutube.GetVideo(videoID)
	if err != nil {
		panic(err)
	}

	formats := video.Formats.WithAudioChannels() // only get videos with audio
	stream, _, err := clientYoutube.GetStream(video, &formats[0])
	if err != nil {
		return "", err
	}

	fileName := video.Title + fmt.Sprintf("%d.mp4", time.Now().UnixNano())
	fileName = strings.ReplaceAll(strings.ToLower(fileName), " ", "_")

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return "", err
	}
	return fileName, nil
}

func (c *Client) convertMp4ToMp3(fileName string) ([]byte, error) {
	// Convert video to MP3

	mp3File := fmt.Sprintf("%s.mp3", fileName)
	//ffmpeg -i input_video.mp4 -vn -acodec libmp3lame -b:a 128k -ar 44100 -ac 2 -threads 4 -preset ultrafast output_audio.mp3

	cmd := exec.Command("ffmpeg", "-hwaccel", "auto", "-i", fileName, "-vn", "-acodec", "libmp3lame", "-b:a", "96k", "-ar", "44100", "-ac", "2", "-s", "640x360", "-threads", "8", "-preset", "ultrafast", mp3File)

	//cmd := exec.Command("ffmpeg", "-i", fileName, "-vn", "-acodec", "libmp3lame",
	//	"-b:a", "128k", "-ar", "44100", "-ac", "2", "-threads", "4", "-preset ultrafast", mp3File)

	fmt.Println(cmd.String())
	err := cmd.Run()

	if err != nil {
		return nil, err
	}

	mp3Bytes, err := os.ReadFile(fmt.Sprintf("%s.mp3", fileName))
	if err != nil {
		return nil, err
	}
	return mp3Bytes, nil

}
