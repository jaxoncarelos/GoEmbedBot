package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	// get token from .env
	godotenv.Load()
	Token := os.Getenv("TOKEN")
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	<-sc

	dg.Close()
}

const (
	Twitter   = 0
	Tiktok    = 1
	Reddit    = 2
	Instagram = 3
)

var regex map[int]string = map[int]string{
	Twitter:   `https:\/\/(?:www\.)?(twitter|x)\.com\/.+\/status(?:es)?\/(\d+)(?:.+ )?`,
	Tiktok:    `https?://(?:www\.|vm\.|vt\.)?tiktok\.com/.+(?: )?`,
	Reddit:    `https?://(?:(?:old\.|www\.)?reddit\.com|v\.redd\.it)/.+(?: )?`,
	Instagram: `https?:\/\/(?:www\.)?instagram\.com\/[a-zA-Z0-9_]+\/?(?:\?igshid=[a-zA-Z0-9_]+)?`,
}

func ShouldBeSpoilered(content string) bool {
	pattern := `^([|]{2}).*$1$`
	if match, _ := regexp.MatchString(pattern, content); match {
		return true
	}
	return false
}
func IsValidUrl(url string) int {
	for i, v := range regex {
		pattern := regexp.MustCompile(v)
		if match := pattern.MatchString(url); match {
			return i
		}
	}
	return -1
}
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
func DownloadVideoFile(url string, should_be_spoiled bool) (string, string, error) {
	outPath := "output.mp4"
	if should_be_spoiled {
		outPath = "SPOILER_output.mp4"
	}
	if fileExists(outPath) {
		os.Remove(outPath)
	}
	cmd := exec.Command("yt-dlp", "-f", "bestvideo[filesize<30MB]+bestaudio[filesize<10mb]/best/bestvideo+bestaudio", "-S", "vcodec:h264", "--merge-output-format", "mp4", "--ignore-config", "--verbose", "--no-playlist", "--no-warnings", "-o", outPath, url)
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
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	content := m.Message.Content
	if strings.HasPrefix(content, "!!") {
		log.Printf("Did no embed on %s\n", content)
		return
	}
	should_be_spoiled := ShouldBeSpoilered(content)
	isValid := IsValidUrl(content)
	if isValid < 0 {
		return
	}
	fmt.Println("Message Created")
	fmt.Printf("Author: %s\n", m.Author.Username)
	fmt.Printf("Message: %s\n", m.Content)
	// delete everything in the string thats not a match
	checkRegex := regexp.MustCompile(regex[isValid])
	content = checkRegex.FindString(content)
	switch isValid {
	case Twitter:
		cmd := exec.Command("yt-dlp", "-g", "-f", "best[ext=mp4]", content)
		output, err := cmd.Output()
		if err != nil {
			log.Printf("Error getting twitter video: %s\n", err)
			return
		}
		s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Reference:       m.Reference(),
			AllowedMentions: &discordgo.MessageAllowedMentions{},
			Content:         fmt.Sprintf("[Twitter Video](%s)", output),
		})
	default:
		output, outPath, err := DownloadVideoFile(content, should_be_spoiled)
		if err != nil {
			log.Printf("Error downloading video: %s\n", err)
			if isValid == Tiktok {
				s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			}
			return
		}
		if output == "" {
			return
		}
		bytes, err := os.ReadFile(outPath)
		if err != nil {
			log.Printf("Error opening file: %s\n", err)
			return
		}
		s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Reference:       m.Reference(),
			AllowedMentions: &discordgo.MessageAllowedMentions{},
			File: &discordgo.File{
				Name:        outPath,
				Reader:      strings.NewReader(string(bytes)),
				ContentType: "video/mp4",
			},
		})
	}

}
