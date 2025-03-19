package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot からのメッセージは無視する
	if m.Author.Bot {
		return
	}
	log.Printf("Message from %s: %s", m.Author.Username, m.Content)
	log.Printf("Channel ID: %s", m.ChannelID)

	if sendChannelID == "" {
		sendChannelID = m.ChannelID
	}
	if m.Content == "ping" {
		s.ChannelMessageSend(sendChannelID, "Pong")
	}

	if strings.HasPrefix(m.Content, "sendChannel") {
		sendChannelID = channelName2ID(s, m.GuildID, m.Content[12:])
		if sendChannelID == "" {
			log.Println("Channel not found")
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Send channel seted")
	}
}

func channelName2ID(s *discordgo.Session, guildID string, channelName string) string {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		log.Println("Error getting channels:", err)
		return sendChannelID
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID
		}
	}

	return sendChannelID
}
