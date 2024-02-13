package main

import (
	"fmt"
	"os"
  "jaxon/embedbot/handler"
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
	dg.AddHandler(handler.MessageCreate)

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

