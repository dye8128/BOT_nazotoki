package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var newEventPrefix = "イベント「"
var newEventSuffix = "」を作成しました\n参加した方はスタンプを押してください"

func onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Member.User.Bot {
		return
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		})
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
			} else if sendChannelID == i.ChannelID {
				sendMessage(s, i, "Select another channel for event creation")
				return
			}

			eventName := strVals[0].StringValue()
			eventLabel := strings.ToUpper(strVals[1].StringValue())
			description := ""
			if len(strVals) > 2 {
				description = strVals[2].StringValue()
			}
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
			// 親チャンネルが存在する場合はその下に作成
			labelID, err := labelName2ID(s, i.GuildID, eventLabel)
			if err != nil {
				raiseError(s, i, "Error getting parent channel", err)
				return
			}
			// 親チャンネルが存在しない場合は新規作成
			if labelID == "" {
				labelData := discordgo.GuildChannelCreateData{
					Name: eventLabel,
					Type: discordgo.ChannelTypeGuildCategory,
				}
				label, err := s.GuildChannelCreateComplex(i.GuildID, labelData)
				if err != nil {
					raiseError(s, i, "Error creating parent channel", err)
					return
				}
				labelID = label.ID
			}

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
				Topic: description,
				ParentID: labelID,
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
			message := newEventPrefix + eventName + newEventSuffix
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
		case "delete_event":
			eventName := strVals[0].StringValue()
			channelID, err := channelName2IDwithGuildID(s, i.GuildID, eventName)
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			_, err = s.ChannelDelete(channelID)
			if err != nil {
				raiseError(s, i, "Error deleting channel", err)
				return
			}
			roleID, err := roleName2ID(s, i.GuildID, eventName)
			if err != nil {
				raiseError(s, i, "Error getting role", err)
				return
			}
			err = s.GuildRoleDelete(i.GuildID, roleID)
			if err != nil {
				raiseError(s, i, "Error deleting role", err)
				return
			}
			message := fmt.Sprintf("イベント「%s」を削除しました", eventName)
			sendMessage(s, i, message)
		case "move_event":
			eventName := strVals[0].StringValue()
			eventLabel := strings.ToUpper(strVals[1].StringValue())
			channelID, err := channelName2IDwithGuildID(s, i.GuildID, eventName)
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			channel, err := s.Channel(channelID)
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			// レーベルから親チャンネルIDを取得
			labelID, err := labelName2ID(s, i.GuildID, eventLabel)
			if err != nil {
				raiseError(s, i, "Error getting parent channel", err)
				return
			}
			if labelID == "" {
				raiseError(s, i, "Parent channel not found", nil)
				return
			}
			// チャンネルの親チャンネルを変更
			pastLabelID := getParentID(s, channelID)
			channelEditData := discordgo.ChannelEdit{
				ParentID: labelID,
			}
			_, err = s.ChannelEditComplex(channel.ID, &channelEditData)
			if err != nil {
				raiseError(s, i, "Error editing channel", err)
				return
			}
			if len(getChildIDs(s, i.GuildID, pastLabelID)) == 0 {
				_, err = s.ChannelDelete(pastLabelID)
				if err != nil {
					raiseError(s, i, "Error deleting label", err)
					return
				}
			}
			message := fmt.Sprintf("イベント「%s」を「%s」に移動しました", eventName, eventLabel)
			sendMessage(s, i, message)
		}
	}
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
	log.Println(message.Content)

	if strings.HasPrefix(message.Content, newEventPrefix) && strings.HasSuffix(message.Content, newEventSuffix) {
		eventName := message.Content[len(newEventPrefix):len(message.Content)-len(newEventSuffix)]
		log.Println(eventName)
		roleID, err := roleName2ID(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting role:", err)
			return
		}
		err = s.GuildMemberRoleAdd(r.GuildID, r.UserID, roleID)
		if err != nil {
			log.Println("Error adding role:", err)
			return
		}
		channelID, err := channelName2IDwithGuildID(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting channel:", err)
			return
		}
		message := fmt.Sprintf("<@%s> がこのイベントに参加しました！", r.UserID)
		s.ChannelMessageSend(channelID, message)
	}
}

func reactionRemove(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	if r.UserID == s.State.User.ID {
		return
	}
	log.Printf("Reaction removed by %s", r.UserID)
	message, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Println("Error getting message:", err)
		return
	}
	log.Println(message.Content)

	if strings.HasPrefix(message.Content, newEventPrefix) && strings.HasSuffix(message.Content, newEventSuffix) {
		eventName := message.Content[len(newEventPrefix):len(message.Content)-len(newEventSuffix)]
		log.Println(eventName)
		roleID, err := roleName2ID(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting role:", err)
			return
		}
		err = s.GuildMemberRoleRemove(r.GuildID, r.UserID, roleID)
		if err != nil {
			log.Println("Error removing role:", err)
			return
		}
		channelID, err := channelName2IDwithGuildID(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting channel:", err)
			return
		}
		message := fmt.Sprintf("<@%s> がこのイベントから離脱しました", r.UserID)
		s.ChannelMessageSend(channelID, message)
	}
}

func sendMessage(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
		Content: message,
	})
}

func raiseError(s *discordgo.Session, i *discordgo.InteractionCreate, errMessage string, err error) {
	log.Println(errMessage, err)
	sendMessage(s, i, errMessage)
}

func channelName2ID(s *discordgo.Session, i *discordgo.InteractionCreate, channelName string) (string, error) {
	// here は送信されたチャンネル
	if channelName == "here" {
		return i.ChannelID, nil
	}

	return channelName2IDwithGuildID(s, i.GuildID, channelName)
}

func channelName2IDwithGuildID(s *discordgo.Session, guildID string, channelName string) (string, error) {
	// チャンネルリンク形式
	if strings.HasPrefix(channelName ,"https://discord") {
		channelName = strings.Split(channelName, "/")[5]
	}

	// チャンネル名の文字列
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return "", err
	}
	for _, channel := range channels {
		if channel.Name == strings.ToLower(channelName) {
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

func roleName2ID(s *discordgo.Session, guildID string, roleName string) (string, error) {
	// ロールメンション形式
	if strings.HasPrefix(roleName, "<@&") {
		return strings.TrimRight(strings.TrimLeft(roleName, "<@&"), ">"), nil
	}

	// ロール名の文字列
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return "", err
	}
	for _, role := range roles {
		if role.Name == roleName {
			return role.ID, nil
		}
	}

	_, err = s.State.Role(guildID, roleName)
	if err != nil {
		return "", err
	}
	return roleName, nil
}

func getParentID(s *discordgo.Session, channelID string) string {
	channel, err := s.Channel(channelID)
	if err != nil {
		log.Println("Error getting channel:", err)
		return ""
	}

	return channel.ParentID
}

func getParentIDs(s *discordgo.Session, guildID string) map[string]string {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		log.Println("Error getting channels:", err)
		return nil
	}

	parentIDkeys := make(map[string]struct{})
	for _, channel := range channels {
		if channel.ParentID != "" {
			parentIDkeys[channel.ParentID] = struct{}{}
		}
	}
	parentIDs := make(map[string]string)
	for parentID := range parentIDkeys {
		parentIDs[parentID] = parentID
	}

	return parentIDs
}

func labelName2ID(s *discordgo.Session, guildID string, labelName string) (string, error) {
	for _, parentID := range getParentIDs(s, guildID) {
		parent, err := s.Channel(parentID)
		if err != nil {
			return "", err
		}
		if parent.Name == labelName {
			return parentID, nil
		}
	}
	return "", nil
}

func getChildIDs(s *discordgo.Session, guildID string, parentID string) []string {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		log.Println("Error getting channels:", err)
		return nil
	}

	childIDs := make([]string, 0)
	for _, channel := range channels {
		if channel.ParentID == parentID {
			childIDs = append(childIDs, channel.ID)
		}
	}

	return childIDs
}