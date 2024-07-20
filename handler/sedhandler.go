package handler

import (
	"errors"
	"strings"
)

func HandleMessage(messages []string, content string) (string, error) {
	split := strings.Split(content, "/")
	if len(split) < 3 {
		return "", errors.New("Lack of arguments")
	}
	for i := 10; i > 0; i-- {
		if strings.Contains(messages[i], split[1]) {
			return strings.Replace(messages[i], split[1], split[2], 1), nil
		}
	}
	return "", errors.New("No match found")
}
