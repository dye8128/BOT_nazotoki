package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.Bot {
		return
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		data := i.ApplicationCommandData()
		strVals := data.Options
		switch data.Name {
		case "send_channel":
			sendChannel, err := channelName2ID(s, i, strVals[0].StringValue())
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			sendChannelIDs[i.GuildID] = sendChannel
			message	:= fmt.Sprintf("招待チャンネルを <#%s> に設定しました", sendChannel)
			sendMessage(s, i, message)
		case "new_event":
			sendChannelID, exists := sendChannelIDs[i.GuildID]
			if !exists {
				sendMessage(s, i, "Channel not set")
				return
			}

			eventName := strVals[0].StringValue()
			// 既にロールが存在するか確認
			roles, err := s.GuildRoles(i.GuildID)
			if err != nil {
				raiseError(s, i, "Error getting roles", err)
				return
			}
			for _, role := range roles {
				if role.Name == eventName {
					sendMessage(s, i, "Role already exists")
					return
				}
			}
			// 既にチャンネルが存在するか確認
			channels, err := s.GuildChannels(i.GuildID)
			if err != nil {
				raiseError(s, i, "Error getting channels", err)
				return
			}
			for _, channel := range channels {
				if channel.Name == eventName {
					sendMessage(s, i, "Channel already exists")
					return
				}
			}
			// ロール作成
			roleData := &discordgo.RoleParams{
				Name: eventName,
			}
			role, err := s.GuildRoleCreate(i.GuildID, roleData)
			if err != nil {
				raiseError(s, i, "Error creating role", err)
				return
			}
			// チャンネル作成
			// @everyone の閲覧権限
			permissionEveryone := &discordgo.PermissionOverwrite{
				ID:   i.GuildID, // @everyone はサーバーID
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			}
			// ロールの閲覧権限
			permissionRole := &discordgo.PermissionOverwrite{
				ID:    role.ID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel,
			}
			channelData := discordgo.GuildChannelCreateData{
				Name: strVals[0].StringValue(),
				Type: discordgo.ChannelTypeGuildText,
				ParentID: getParentID(s, sendChannelID),
				PermissionOverwrites: []*discordgo.PermissionOverwrite{
					permissionEveryone,
					permissionRole,
				},
			}
			_, err = s.GuildChannelCreateComplex(i.GuildID, channelData)
			if err != nil {
				raiseError(s, i, "Error creating channel", err)
				return
			}
			message := fmt.Sprintf("イベント %s を作成しました\n参加した方はスタンプを押してください", eventName)
			s.ChannelMessageSend(sendChannelID, message)
			// 最新のメッセージを取得
			messages, err := s.ChannelMessages(sendChannelID, 1, "", "", "")
			if err != nil {
				raiseError(s, i, "Error getting messages", err)
				return
			}
			// メッセージにリアクションを追加
			err = s.MessageReactionAdd(sendChannelID, messages[0].ID, "👍")
			if err != nil {
				log.Println("Error adding reaction:", err)
				s.ChannelMessageSend(sendChannelID, "Error adding reaction")
				s.ChannelMessageDelete(sendChannelID, messages[0].ID)
				return
			}
			// 自分にロールを付与
			err = s.GuildMemberRoleAdd(i.GuildID, s.State.User.ID, role.ID)
			if err != nil {
				raiseError(s, i, "Error adding role", err)
				s.ChannelMessageDelete(sendChannelID, messages[0].ID)
				return
			}
			message = fmt.Sprintf("イベント「%s」の作成が正常に行われました", eventName)
			sendMessage(s, i, message)
		}
	}
}

func sendMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}

func raiseError(s *discordgo.Session, i *discordgo.InteractionCreate, errMessage string, err error) {
	log.Println(errMessage, err)
	sendMessage(s, i, errMessage)
}

func channelName2ID(s *discordgo.Session, i *discordgo.InteractionCreate, channelName string) (string, error) {
	// here が指定された場合はそのチャンネル
	if channelName == "here" {
		return i.ChannelID, nil
	}

	// チャンネルリンク形式
	if strings.HasPrefix(channelName ,"https://discord") {
		channelName = strings.Split(channelName, "/")[5]
	}

	// チャンネル名の文字列
	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		return "", err
	}
	for _, channel := range channels {
		if channel.Name == channelName {
			channelName = channel.ID
			break
		}
	}
	_, err = s.Channel(channelName)
	if err != nil {
		return "", err
	}
	return channelName, nil
}

func getParentID(s *discordgo.Session, channelID string) string {
	channel, err := s.Channel(channelID)
	if err != nil {
		log.Println("Error getting channel:", err)
		return ""
	}

	return channel.ParentID
}