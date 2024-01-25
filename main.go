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
	fmt.Println("Hello, World!")
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

var regex map[string]string = map[string]string{
	"twitter":   `https:\/\/(?:www\.)?(twitter|x)\.com\/.+\/status(?:es)?\/(\d+)(?:.+ )?`,
	"tiktok":    `https?://(?:www\.|vm\.|vt\.)?tiktok\.com/.+(?: )?`,
	"reddit":    `https?://(?:(?:old\.|www\.)?reddit\.com|v\.redd\.it)/.+(?: )?`,
	"instagram": `https?:\/\/(?:www\.)?instagram\.com\/[a-zA-Z0-9_]+\/?(?:\?igshid=[a-zA-Z0-9_]+)?`,
}

var urlRegex string = `(https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z]{2,}(\.[a-zA-Z]{2,})(\.[a-zA-Z]{2,})?\/[a-zA-Z0-9]{2,}|((https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z]{2,}(\.[a-zA-Z]{2,})(\.[a-zA-Z]{2,})?)|(https:\/\/www\.|http:\/\/www\.|https:\/\/|http:\/\/)?[a-zA-Z0-9]{2,}\.[a-zA-Z0-9]{2,}\.[a-zA-Z0-9]{2,}(\.[a-zA-Z0-9]{2,})?`

func should_be_spoilered(content string) bool {
	pattern := `^([|]{2}).*$1$`
	if match, _ := regexp.MatchString(pattern, content); match {
		return true
	}
	return false
}
func is_valid_url(url string) string {
	for i, v := range regex {
		pattern := regexp.MustCompile(v)
		if match := pattern.MatchString(url); match {
			return i
		}
	}
	return ""
}
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
func download_video_file(url string, should_be_spoiled bool) (string, string) {
	outPath := "output.mp4"
	if should_be_spoiled {
		outPath = "SPOILER_output.mp4"
	}
	if fileExists(outPath) {
		os.Remove(outPath)
	}
	// ["yt-dlp",
	//                     "-f", "bestvideo[filesize<6MB]+bestaudio[filesize<2MB]/best/bestvideo+bestaudio",
	//                     "-S", "vcodec:h264",
	//                     "--merge-output-format", "mp4",
	//                     "--ignore-config",
	//                     "--verbose",
	//                     "--cookies", "cookies.txt" if "instagram" in content else ""
	//                     "--no-playlist",
	//                     "--no-warnings", '-o', outPath, content,
	//                     ]
	cmd := exec.Command("yt-dlp", "-f", "bestvideo[filesize<30MB]+bestaudio[filesize<10mb]/best/bestvideo+bestaudio", "-S", "vcodec:h264", "--merge-output-format", "mp4", "--ignore-config", "--verbose", "--no-playlist", "--no-warnings", "-o", "1"+outPath, url)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("%s\n", stderr.String())
		return "", ""
	}
	cmd = exec.Command("ffmpeg", "-i", "1"+outPath, "-y", "-c:v", "libx264", "-crf", "23", "-preset", "ultrafast", "-c:a", "copy", outPath)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("%s\n", stderr.String())
		return "", ""
	}
	os.Remove("1" + outPath)
	output := out.String()
	return output, outPath
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
	should_be_spoiled := should_be_spoilered(content)
	is_valid := is_valid_url(content)
	if is_valid == "" {
		return
	}
	fmt.Println("Message Created")
	fmt.Printf("Author: %s\n", m.Author.Username)
	fmt.Printf("Message: %s\n", m.Content)
	// delete everything in the string thats not a match
	checkRegex := regexp.MustCompile(regex[is_valid])
	content = checkRegex.FindString(content)
	if is_valid != "" {
		// })
		output, outPath := download_video_file(content, should_be_spoiled)

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
				Name: outPath,
				// mp4 file reader
				Reader:      strings.NewReader(string(bytes)),
				ContentType: "video/mp4",
			},
		})
	}

}
