package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var sendChannelIDs = make(map[string]string)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	TOKEN := os.Getenv("TOKEN")
	MY_GUILD_ID := os.Getenv("MY_GUILD_ID")
	MY_CHANNEL_ID := os.Getenv("MY_CHANNEL_ID")
	sendChannelIDs[MY_GUILD_ID] = MY_CHANNEL_ID

	dg, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(onInteractionCreate)
	dg.AddHandler(reactionAdd)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	commands := []discordgo.ApplicationCommand{
		{
			Name: 	     "send_channel",
			Description: "Set send channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel",
					Description: "Channel name or link",
					Required:    true,
				},
			},
		},
		{
			Name:        "new_event",
			Description: "Create new event",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Event name",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete_event",
			Description: "Delete event",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "Event name",
					Required:    true,
				},
			},
		},
	}

	for _, command := range commands {
		_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", &command)
		if err != nil {
			log.Fatal(err)
		}
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
