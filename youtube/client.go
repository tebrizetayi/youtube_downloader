package youtube

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type YoutubeDownloader interface {
	DownloadYouTubeMP3(url string) (string, error)
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

func (c *Client) DownloadYouTubeMP3(url string) (string, error) {

	videoID, err := c.extractVideoID(url)
	if err != nil {
		return "", nil
	}
	filename, err := c.download(videoID)
	if err != nil {
		return "", err
	}

	filename, err = c.convertMp4ToMp3(filename)
	if err != nil {
		return "", err
	}

	return filename, nil
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

	fileName := video.Title
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ToLower(fileName)

	file, err := os.Create(fmt.Sprintf("%s.mp4", fileName))
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

func (c *Client) convertMp4ToMp3(fileName string) (string, error) {
	// Convert video to MP3

	mp3File := fmt.Sprintf("%s.mp3", fileName)
	cmd := exec.Command("ffmpeg", "-i", fmt.Sprintf("%s.mp4", fileName), "-vn", "-acodec", "libmp3lame", "-ac", "2", "-ab", "160k", "-ar", "48000", mp3File)
	err := cmd.Run()
	fmt.Println(cmd.String())
	if err != nil {
		return "", err
	}
	return mp3File, nil
}
