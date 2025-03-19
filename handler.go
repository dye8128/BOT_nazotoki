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

	if sendChannelID == "" {
		sendChannelID = m.ChannelID
	}
	if m.Content == "ping" {
		s.ChannelMessageSend(sendChannelID, "Pong")
	}

	if strings.HasPrefix(m.Content, "sendChannel") {
		strVal := strings.Split(m.Content, " ")[1]
		// prefix: sendChannel link {チャンネルリンク} OR sendChannel {チャンネル名}
		if strVal == "link" {
			sendChannelID = strings.Split((strings.Split(m.Content, " ")[2]), "/")[5]
		} else {
			sendChannelID = channelName2ID(s, m.GuildID, strVal)
		}
		if sendChannelID == "" {
			log.Println("Channel not found")
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Send channel seted")
	}

	if strings.HasPrefix(m.Content, "deleteMessage") {
		// prefix: deleteMessage {メッセージリンク}
		// メッセージリンクからメッセージIDを取得
		strVal := strings.Split(m.Content, " ")[1]
		messageID := strings.Split(strVal, "/")[6]
		log.Println("Delete message ID:", messageID)
		err := s.ChannelMessageDelete(m.ChannelID, messageID)
		if err != nil {
			log.Println("Error deleting message:", err)
			s.ChannelMessageSend(m.ChannelID, "Error deleting message")
		}
	}

	if strings.HasPrefix(m.Content, "createRole") {
		// prefix: createRole {ロール名}
		roleName := strings.Split(m.Content, " ")[1]
		roleData := &discordgo.RoleParams{
			Name: roleName,
		}

		if existsRole(s, m.GuildID, roleName) {
			s.ChannelMessageSend(m.ChannelID, "Role already exists")
			return
		}

		role, err := s.GuildRoleCreate(m.GuildID, roleData)
		if err != nil {
			log.Println("Error creating role:", err)
			s.ChannelMessageSend(m.ChannelID, "Error creating role")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Role created: "+role.Name)
	}
}

// チャンネル名からチャンネルIDを取得する
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

func existsRole(s *discordgo.Session, guildID string, roleName string) bool {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		log.Println("Error getting roles:", err)
		return false
	}

	for _, role := range roles {
		if role.Name == roleName {
			return true
		}
	}

	return false
}
