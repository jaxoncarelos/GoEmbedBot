package ContentUtils

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
)

const (
	Twitter = iota
	Tiktok
	Reddit
	Instagram
)

var regex map[int]string = map[int]string{
	Twitter:   `https:\/\/(?:www\.)?(twitter|x)\.com\/.+\/status(?:es)?\/(\d+)(?:.+ )?`,
	Tiktok:    `https?://(?:www\.|vm\.|vt\.)?tiktok\.com/.+(?: )?`,
	Reddit:    `https?://(?:(?:old\.|www\.)?reddit\.com|v\.redd\.it)/.+(?: )?`,
	Instagram: `https?:\/\/(?:www\.)?instagram\.com\/[a-zA-Z0-9_]+\/?(?:\?igshid=[a-zA-Z0-9_]+)?`,
}

func GetRegex(index int) string {
	return regex[index]
}

func ShouldBeSpoilered(content string) bool {
	pattern := `^([|]{2}).*([|]{2})$`
	if match, _ := regexp.MatchString(pattern, content); match {
		return true
	}
	return false
}

func IsValidUrl(url string) (int, error) {
	for i, v := range regex {
		pattern := regexp.MustCompile(v)
		if match := pattern.MatchString(url); match {
			return i, nil
		}
	}
	return -1, errors.New("Invalid URL")
}

func FileExists(filename string) error {
	_, err := os.Stat(filename)
	return err
}

func DownloadTikTokVideo(url string, should_be_spoiled bool) (string, string, error) {
	outPath := "output.mp4"
	if should_be_spoiled {
		outPath = "SPOILER_output.mp4"
	}
	{
		err := FileExists(outPath)
		if err == nil {
			os.Remove(outPath)
		}
	}
	cmd := exec.Command(
		"yt-dlp",
		"-S",
		"vcodec:h264,res:576",
		"-o",
		outPath,
		url,
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("%s\n", stderr.String())
		return "", "", err
	}
	output := out.String()
	return output, outPath, nil
}

func DownloadVideoFile(url string, should_be_spoiled bool) (string, string, error) {
	outPath := "output.mp4"
	if should_be_spoiled {
		outPath = "SPOILER_output.mp4"
	}
	{
		err := FileExists(outPath)
		if err == nil {
			os.Remove(outPath)
		}
	}
	cmd := exec.Command(
		"yt-dlp",
		"-f",
		"bestvideo[filesize<30MB]+bestaudio[filesize<10mb]/best/bestvideo+bestaudio",
		"-S",
		"vcodec:h264",
		"--merge-output-format",
		"mp4",
		"--ignore-config",
		"--verbose",
		"--no-playlist",
		"--no-warnings",
		"-o",
		outPath,
		url,
	)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("%s\n", stderr.String())
		return "", "", err
	}
	output := out.String()
	return output, outPath, nil
}
