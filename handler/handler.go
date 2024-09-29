package handler

import (
	"fmt"
	"jaxon/embedbot/ContentUtils"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var sedHistory map[string][]MessageHandler

type MessageHandler struct {
	User    string
	Content string
}

func MessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	content := m.Message.Content
	{
		if sedHistory == nil {
			sedHistory = map[string][]MessageHandler{}
		}
		var user string
		if strings.HasPrefix(content, "sed/") {
			if m.Member.Nick != "" {
				user = m.Member.Nick
			} else {
				user = m.Author.Username
			}
			newContent, err := HandleMessage(sedHistory[m.ChannelID], content)
			if err != nil {
				log.Printf("Error handling message: %s\n", err)
				return
			}
			s.ChannelMessageSend(m.ChannelID, newContent)
			return
		}
		sedHistory[m.ChannelID] = append(sedHistory[m.ChannelID], MessageHandler{
			User:    user,
			Content: content,
		})
		if len(sedHistory[m.ChannelID]) > 30 {
			sedHistory[m.ChannelID] = sedHistory[m.ChannelID][1:]
		}

	}
	if strings.HasPrefix(content, "!!") || strings.HasPrefix(content, ".dl") {
		log.Printf("Did no embed on %s\n", content)
		return
	}
	should_be_spoiled := ContentUtils.ShouldBeSpoilered(content)
	isValid, err := ContentUtils.IsValidUrl(content)
	if err != nil {
		return
	}
	fmt.Println("Message Created")
	fmt.Printf("Author: %s\n", m.Author.Username)
	fmt.Printf("Message: %s\n", m.Content)
	checkRegex := regexp.MustCompile(ContentUtils.GetRegex(isValid))
	content = checkRegex.FindString(content)
	switch isValid {
	case ContentUtils.Twitter:
		cmd := exec.Command(
			"yt-dlp",
			"-g",
			"-f",
			"best[ext=mp4]",
			strings.Replace(content, "https://x.com", "https://twitter.com", 1),
		)
		output, err := cmd.Output()
		if err != nil {
			log.Printf("Error getting video: %s\n", err)
			return
		}
		toSend := fmt.Sprintf("[Video](%s)", output)
		if should_be_spoiled {
			toSend = fmt.Sprintf("||%s||", toSend)
		}
		s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Reference:       m.Reference(),
			AllowedMentions: &discordgo.MessageAllowedMentions{},
			Content:         toSend,
		})

	default:
		output, outPath, err := ContentUtils.DownloadVideoFile(content, should_be_spoiled)
		if err != nil {
			log.Printf("Error downloading video: %s\n", err)
			if isValid == ContentUtils.Tiktok {
				s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			}
			return
		}
		if output == "" {
			return
		}
		bytes, err := os.ReadFile(outPath)
		defer os.Remove(outPath)
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
