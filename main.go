package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	TOKEN := os.Getenv("TOKEN")

	discord, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	discord.AddHandler(onMessageCreate)

	err = discord.Open()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot is running")

	stopBot := make(chan os.Signal, 1)
	signal.Notify(stopBot, os.Interrupt)
	<-stopBot

	log.Println("Bot is stopping")
	
	err = discord.Close()
	if err != nil {
		log.Fatal(err)
	}
}
