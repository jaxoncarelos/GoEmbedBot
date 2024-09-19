package handler

import (
	"errors"
	"strings"
)

func HandleMessage(messages []MessageHandler, content string) (string, error) {
	split := strings.Split(content, "/")
	if len(split) < 3 {
		return "", errors.New("Lack of arguments")
	}
	for i := len(messages) - 1; i > 0; i-- {
		if strings.Contains(messages[i].Content, split[1]) {
			return messages[i].User + ": " + strings.Replace(messages[i].Content, split[1], split[2], 1), nil
		}
	}
	// one in this file
	return "", errors.New("No match found")
}
