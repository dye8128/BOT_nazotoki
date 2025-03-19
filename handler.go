package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Bot からのメッセージは無視する
	if m.Author.Bot {
		return
	}
	log.Printf("Message from %s: %s", m.Author.Username, m.Content)

	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong")
	}
}