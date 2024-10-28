package ContentUtils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

const (
	Twitter = iota
	Tiktok
	Reddit
	Instagram
	Facebook
)

type MessageEdit struct {
	Content         *string                           `json:"content,omitempty"`
	Components      []discordgo.MessageComponent      `json:"components,omitempty"`
	Embeds          []*discordgo.MessageEmbed         `json:"embeds,omitempty"`
	AllowedMentions *discordgo.MessageAllowedMentions `json:"allowed_mentions,omitempty"`
	Flags           discordgo.MessageFlags            `json:"flags,omitempty"`
	// Files to append to the message
	Files []*discordgo.File `json:"-"`
	// Overwrite existing attachments
	Attachments *[]*discordgo.MessageAttachment `json:"attachments,omitempty"`

	ID      string
	Channel string

	// TODO: Remove this when compatibility is not required.
	Embed *discordgo.MessageEmbed `json:"-"`
}

var (
	// Marshal defines function used to encode JSON payloads
	Marshal func(v interface{}) ([]byte, error) = json.Marshal
	// Unmarshal defines function used to decode JSON payloads
	Unmarshal func(src []byte, v interface{}) error = json.Unmarshal
)

func custom_request(s *discordgo.Session, method, urlStr, contentType string, b []byte, bucketID string, sequence int, options ...discordgo.RequestOption) (response []byte, err error) {
	if bucketID == "" {
		bucketID = strings.SplitN(urlStr, "?", 2)[0]
	}
	return s.RequestWithLockedBucket(method, urlStr, contentType, b, s.Ratelimiter.LockBucket(bucketID), sequence, options...)
}
func unmarshal(data []byte, v interface{}) error {
	err := Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("%w: %s", "123", err)
	}

	return nil
}
func CustomChannelMessageEditComplex(s *discordgo.Session, m *MessageEdit, options ...discordgo.RequestOption) (st *discordgo.Message, err error) {
	// TODO: Remove this when compatibility is not required.
	if m.Embed != nil {
		if m.Embeds == nil {
			m.Embeds = []*discordgo.MessageEmbed{m.Embed}
		} else {
			err = fmt.Errorf("cannot specify both Embed and Embeds")
			return
		}
	}

	for _, embed := range m.Embeds {
		if embed.Type == "" {
			embed.Type = "rich"
		}
	}

	endpoint := discordgo.EndpointChannelMessage(m.Channel, m.ID)

	var response []byte
	if len(m.Files) > 0 {
		contentType, body, encodeErr := discordgo.MultipartBodyWithJSON(m, m.Files)
		if encodeErr != nil {
			return st, encodeErr
		}
		response, err = custom_request(s, "PATCH", endpoint, contentType, body, discordgo.EndpointChannelMessage(m.Channel, ""), 0, options...)
	} else {
		response, err = s.RequestWithBucketID("PATCH", endpoint, m, discordgo.EndpointChannelMessage(m.Channel, ""), options...)
	}
	if err != nil {
		return
	}

	err = unmarshal(response, &st)
	return
}

var regex map[int]string = map[int]string{
	Twitter:   `https:\/\/(?:www\.)?(twitter|x)\.com\/.+\/status(?:es)?\/(\d+)(?:.+ )?`,
	Tiktok:    `https?://(?:www\.|vm\.|vt\.)?tiktok\.com/.+(?: )?`,
	Reddit:    `https?://(?:(?:old\.|www\.)?reddit\.com|v\.redd\.it)/.+(?: )?`,
	Instagram: `https?:\/\/(?:www\.)?instagram\.com\/[a-zA-Z0-9_]+\/?(?:\?igshid=[a-zA-Z0-9_]+)?`,
	Facebook:  `https?:\/\/(?:www\.)?facebook\.com\/(reel)\/[a-zA-Z0-9_]+\/?`,
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

func DownloadVideoFile(url string, should_be_spoiled bool) (string, string, error) {
	// generate random constant to put in filename
	fileName := uuid.New()
	outPath := fmt.Sprintf("%s.mp4", fileName)
	if should_be_spoiled {
		outPath = fmt.Sprintf("%d_spoiler.mp4", fileName)
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
