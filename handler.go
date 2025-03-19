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
		// prefix: sendChannel {チャンネルリンク} OR sendChannel {チャンネル名}
		if strings.HasPrefix(strVal, "https://discordapp.com/channels/") {
			sendChannelID = strings.Split(strVal, "/")[5]
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

	if strings.HasPrefix(m.Content, "deleteRole") {
		// prefix: deleteRole {ロール名} OR deleteRole {ロールメンション}
		strVal := strings.Split(m.Content, " ")[1]
		var roleID string
		if strings.HasPrefix(strVal, "<@&") {
			roleID = strings.TrimRight(strings.TrimLeft(strVal, "<@&"), ">")
		} else {
			roleID = roleName2ID(s, m.GuildID, strVal)
		}

		if roleID == "" {
			log.Println("Role not found")
			s.ChannelMessageSend(m.ChannelID, "Role not found")
			return
		}

		err := s.GuildRoleDelete(m.GuildID, roleID)
		if err != nil {
			log.Println("Error deleting role:", err)
			s.ChannelMessageSend(m.ChannelID, "Error deleting role")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Role deleted")
	}

	if(strings.HasPrefix(m.Content, "createChannel")) {
		// prefix: createChannel {チャンネル名}
		channelName := strings.Split(m.Content, " ")[1]
		if(existsChannel(s, m.GuildID, channelName)) {
			s.ChannelMessageSend(m.ChannelID, "Channel already exists")
			return
		}

		s.GuildChannelCreate(m.GuildID, channelName, discordgo.ChannelTypeGuildText)
		s.ChannelMessageSend(m.ChannelID, "Channel created")
	}

	if(strings.HasPrefix(m.Content, "deleteChannel")) {
		// prefix: deleteChannel {チャンネル名} OR deleteChannel {チャンネルリンク}
		strVal := strings.Split(m.Content, " ")[1]
		var channelID string
		if strings.HasPrefix(strVal, "https://discordapp.com/channels/") {
			channelID = strings.Split(strVal, "/")[5]
		} else {
			channelID = channelName2ID(s, m.GuildID, strVal)
		}

		if channelID == "" {
			log.Println("Channel not found")
			s.ChannelMessageSend(m.ChannelID, "Channel not found")
			return
		}
		log.Println("Delete channel ID:", channelID)

		_, err := s.ChannelDelete(channelID)
		if err != nil {
			log.Println("Error deleting channel:", err)
			s.ChannelMessageSend(m.ChannelID, "Error deleting channel")
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Channel deleted")
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

// ロール名からロールIDを取得する
func roleName2ID(s *discordgo.Session, guildID string, roleName string) string {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		log.Println("Error getting roles:", err)
		return ""
	}

	for _, role := range roles {
		if role.Name == roleName {
			return role.ID
		}
	}

	return ""
}

func existsChannel(s *discordgo.Session, guildID string, channelName string) bool {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		log.Println("Error getting channels:", err)
		return false
	}

	for _, channel := range channels {
		if channel.Name == channelName {
			return true
		}
	}

	return false
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
