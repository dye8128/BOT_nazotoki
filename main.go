package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var sendChannelID string

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	TOKEN := os.Getenv("TOKEN")

	dg, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	sendChannelID = ""
	dg.AddHandler(onMessageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is running")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Bot is stopping")

	err = dg.Close()
	if err != nil {
		log.Fatal(err)
	}
}
