package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// エラーは英語、成功メッセージは日本語で返す
func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Bot からのメッセージは無視する
	if i.Member.User.Bot {
		return
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		data := i.ApplicationCommandData()
		strVals := data.Options
		if data.Name == "send_channel" {
			if len(strVals) != 1 {
				valuesLengthError(s, i)
				return
			}
			sendChannel := channelName2ID(s, i, strVals[0].StringValue())
			_, err := s.Channel(sendChannel)
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			sendChannelIDs[i.GuildID] = sendChannel
			message := fmt.Sprintf("ログ送信チャンネルを https://discordapp.com/channels/%s/%s に設定しました", i.GuildID, sendChannel)
			sendMessage(s, i, message)
			return
		}

		sendChannelID, exists := sendChannelIDs[i.GuildID]
		if !exists {
			sendMessage(s, i, "Channel not set")
			return
		}

		switch data.Name {
		case "ping":
			sendMessage(s, i, "Pong")
		case "delete_message":
			if len(strVals) != 1 {
				valuesLengthError(s, i)
				return
			}
			messageID := messageLink2ID(strVals[0].StringValue())
			err := s.ChannelMessageDelete(sendChannelID, messageID)
			if err != nil {
				raiseError(s, i, "Error deleting message", err)
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "該当メッセージを削除しました",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		case "create_role":
			if len(strVals) != 1 {
				valuesLengthError(s, i)
				return
			}
			roleData := &discordgo.RoleParams{
				Name: strVals[0].StringValue(),
			}
			// 既にロールが存在するか確認
			roles, err := s.GuildRoles(i.GuildID)
			if err != nil {
				raiseError(s, i, "Error getting roles", err)
				return
			}
			for _, role := range roles {
				if role.Name == roleData.Name {
					sendMessage(s, i, "Role already exists")
					return
				}
			}
			// ロール作成
			role, err := s.GuildRoleCreate(i.GuildID, roleData)
			if err != nil {
				raiseError(s, i, "Error creating role", err)
				return
			} else {
				message := fmt.Sprintf("ロール %s を作成しました", role.Name)
				sendMessage(s, i, message)
			}
		case "delete_role":
			if len(strVals) != 1 {
				valuesLengthError(s, i)
				return
			}
			roleID := roleName2ID(s, i.GuildID, strVals[0].StringValue())
			err := s.GuildRoleDelete(i.GuildID, roleID)
			if err != nil {
				raiseError(s, i, "Error deleting role", err)
				return
			} else {
				message := "ロールを削除しました"
				sendMessage(s, i, message)
			}
		case "create_channel":
			if len(strVals) != 1 {
				valuesLengthError(s, i)
				return
			}
			channelData := discordgo.GuildChannelCreateData{
				Name: strVals[0].StringValue(),
				Type: discordgo.ChannelTypeGuildText,
				ParentID: getParentID(s, i.ChannelID),
			}
			// 既にチャンネルが存在するか確認
			channels, err := s.GuildChannels(i.GuildID)
			if err != nil {
				raiseError(s, i, "Error getting channels", err)
				return
			}
			for _, channel := range channels {
				if channel.Name == channelData.Name {
					sendMessage(s, i, "Channel already exists")
					return
				}
			}
			// チャンネル作成
			channel, err := s.GuildChannelCreateComplex(i.GuildID, channelData)
			if err != nil {
				raiseError(s, i, "Error creating channel", err)
				return
			} else {
				message := fmt.Sprintf("https://discordapp.com/channels/%s/%s を作成しました", i.GuildID, channel.ID)
				sendMessage(s, i, message)
			}
		case "delete_channel":
			if len(strVals) != 1 {
				valuesLengthError(s, i)
				return
			}
			channelID := channelName2ID(s, i, strVals[0].StringValue())
			_, err := s.ChannelDelete(channelID)
			if err != nil {
				raiseError(s, i, "Error deleting channel", err)
				return
			} else {
				message := "チャンネルを削除しました"
				sendMessage(s, i, message)
			}
		case "to_private":
			if len(strVals) != 1 && len(strVals) != 2 {
				valuesLengthError(s, i)
				return
			}
			roleID := roleName2ID(s, i.GuildID, strVals[0].StringValue())
			var channelID string
			if len(strVals) == 1 {
				channelID = channelName2ID(s, i, "here")
			} else {
				channelID = channelName2ID(s, i, strVals[1].StringValue())
			}

			// チャンネル情報の取得
			channel, err := s.Channel(channelID)
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			// @everyone の閲覧権限
			permissionEveryone := &discordgo.PermissionOverwrite{
				ID:   channel.GuildID, // @everyone はサーバーID
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			}
			// ロールの閲覧権限
			permissionRole := &discordgo.PermissionOverwrite{
				ID:    roleID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel,
			}
			// チャンネルの権限設定
			_, err = s.ChannelEditComplex(channelID, &discordgo.ChannelEdit{
				PermissionOverwrites: []*discordgo.PermissionOverwrite{permissionEveryone, permissionRole},
			})
			if err != nil {
				raiseError(s, i, "Error updating channel permission", err)
				return
			} else {
				message := fmt.Sprintf("https://discordapp.com/channels/%s/%s の閲覧をロール <@&%s> のみに変更しました", i.GuildID, channelID, roleID)
				sendMessage(s, i, message)
			}
		case "add_role":
			if len(strVals) != 1 {
				valuesLengthError(s, i)
				return
			}
			roleID := roleName2ID(s, i.GuildID, strVals[0].StringValue())
			message	:= fmt.Sprintf("ロール <@&%s> を付与します", roleID)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			})
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
				return
			}
			// 自分にロールを付与
			err = s.GuildMemberRoleAdd(i.GuildID, s.State.User.ID, roleID)
			if err != nil {
				raiseError(s, i, "Error adding role", err)
				return
			}
			s.InteractionResponseDelete(i.Interaction)		
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

func valuesLengthError(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "引数の数が異なります",
		},
	})
}

func channelName2ID(s *discordgo.Session, i *discordgo.InteractionCreate, channelName string) string {
	// here は送信されたチャンネル
	if channelName == "here" {
		return i.ChannelID
	}

	// チャンネルリンク形式
	if strings.HasPrefix(channelName ,"https://discord") {
		return strings.Split(channelName, "/")[5]
	}

	// チャンネル名の文字列
	channels, err := s.GuildChannels(i.GuildID)
	if err != nil {
		log.Println("Error getting guild:", err)
		return ""
	}
	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID
		}
	}
	return channelName
}

func messageLink2ID(messageLink string) (string) {
	// メッセージリンク形式
	if strings.HasPrefix(messageLink, "https://discord") {
		return strings.Split(messageLink, "/")[6]
	}

	return messageLink
}

func roleName2ID(s *discordgo.Session, guildID string, roleName string) string {
	// ロールメンション形式
	if strings.HasPrefix(roleName, "<@&") {
		return strings.TrimRight(strings.TrimLeft(roleName, "<@&"), ">")
	}

	// ロール名の文字列
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

	return roleName
}

func getParentID(s *discordgo.Session, channelID string) string {
	channel, err := s.Channel(channelID)
	if err != nil {
		log.Println("Error getting channel:", err)
		return ""
	}

	return channel.ParentID
}

func reactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}
	log.Printf("Reaction added by %s", r.UserID)
	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Println("Error getting message:", err)
		return
	}

	if strings.HasPrefix(message.Content, "ロール <@&") {
		roleID := strings.TrimRight(strings.TrimLeft(message.Content, "ロール <@&"), "> を付与します")
		err := s.GuildMemberRoleAdd(r.GuildID, r.UserID, roleID)
		if err != nil {
			log.Println("Error adding role:", err)
			s.ChannelMessageSend(sendChannelIDs[r.GuildID], "Error adding role")
			return
		}
		message := fmt.Sprintf("<@%s> にロール <@&%s> を付与しました", r.UserID, roleID)
		s.ChannelMessageSend(sendChannelIDs[r.GuildID], message)
	}
}