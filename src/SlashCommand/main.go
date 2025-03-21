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
			Name:        "ping",
			Description: "Pong",
		},
		{
			Name:        "send_channel",
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
			Name:        "delete_message",
			Description: "Delete message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Message link",
					Required:    true,
				},
			},
		},
		{
			Name:        "create_role",
			Description: "Create role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Role name",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete_role",
			Description: "Delete role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Role name",
					Required:    true,
				},
			},
		},
		{
			Name:        "create_channel",
			Description: "Create channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel",
					Description: "Channel name",
					Required:    true,
				},
			},
		},
		{
			Name:        "delete_channel",
			Description: "Delete channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel",
					Description: "Channel name",
					Required:    true,
				},
			},
		},
		{
			Name: 		 "to_private",
			Description: "Channel to private",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Role name",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "channel",
					Description: "Channel name",
					Required:    false,
				},
			},
		},
		{
			Name: 		 "add_role",
			Description: "Add role",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "role",
					Description: "Role name",
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
