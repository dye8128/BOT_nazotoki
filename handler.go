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
			sendChannel, err := getChannelwithInteraction(s, i, strVals[0].StringValue())
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			sendChannelIDs[i.GuildID] = sendChannel.ID
			message := fmt.Sprintf("招待チャンネルを <#%s> に設定しました", sendChannel.ID)
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
			parentName := strings.ToUpper(strVals[1].StringValue())
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
			parent, err := getParent(s, i.GuildID, parentName)
			if err != nil {
				raiseError(s, i, "Error getting parent channel", err)
				return
			}
			// 親チャンネルが存在しない場合は新規作成
			if parent == nil {
				parentData := discordgo.GuildChannelCreateData{
					Name: parentName,
					Type: discordgo.ChannelTypeGuildCategory,
				}
				parent, err = s.GuildChannelCreateComplex(i.GuildID, parentData)
				if err != nil {
					raiseError(s, i, "Error creating parent channel", err)
					return
				}
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
				Name:     eventName,
				Type:     discordgo.ChannelTypeGuildText,
				Topic:    description,
				ParentID: parent.ID,
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
			channel, err := getChannel(s, i.GuildID, eventName)
			eventName = channel.Name
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			_, err = s.ChannelDelete(channel.ID)
			if err != nil {
				raiseError(s, i, "Error deleting channel", err)
				return
			}
			role, err := getRole(s, i.GuildID, eventName)
			if err != nil {
				raiseError(s, i, "Error getting role", err)
				return
			}
			err = s.GuildRoleDelete(i.GuildID, role.ID)
			if err != nil {
				raiseError(s, i, "Error deleting role", err)
				return
			}
			message := fmt.Sprintf("イベント「%s」を削除しました", eventName)
			sendMessage(s, i, message)
		case "move_event":
			eventName := strVals[0].StringValue()
			parentName := strings.ToUpper(strVals[1].StringValue())
			channel, err := getChannel(s, i.GuildID, eventName)
			if err != nil {
				raiseError(s, i, "Error getting channel", err)
				return
			}
			// レーベルから親チャンネルIDを取得
			parent, err := getParent(s, i.GuildID, parentName)
			if err != nil {
				raiseError(s, i, "Error getting parent channel", err)
				return
			}
			if parent == nil {
				raiseError(s, i, "Parent channel not found", nil)
				return
			}
			// チャンネルの親チャンネルを変更
			pastParent, err := getParentfromChild(s, channel.ID)
			if err != nil {
				raiseError(s, i, "Error getting parent channel", err)
				return
			}
			channelEditData := discordgo.ChannelEdit{
				ParentID: parent.ID,
			}
			_, err = s.ChannelEditComplex(channel.ID, &channelEditData)
			if err != nil {
				raiseError(s, i, "Error editing channel", err)
				return
			}
			childIDs, err := getChildIDs(s, i.GuildID, parent.ID)
			if err != nil {
				raiseError(s, i, "Error getting child channels", err)
				return
			}
			if len(childIDs) == 0 {
				_, err = s.ChannelDelete(pastParent.ID)
				if err != nil {
					raiseError(s, i, "Error deleting label", err)
					return
				}
			}
			message := fmt.Sprintf("イベント「%s」を「%s」に移動しました", channel.Name, parent.Name)
			sendMessage(s, i, message)
		}
	}
}

func autocompleteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}

	data := i.ApplicationCommandData()

	// 選択中のオプションを取得
	var focusedOption *discordgo.ApplicationCommandInteractionDataOption
	for _, opt := range data.Options {
		if opt.Focused {
			focusedOption = opt
			break
		}
	}
	if focusedOption == nil {
		return
	}

	Choices := []*discordgo.ApplicationCommandOptionChoice{}

	if focusedOption.Name == "label" {
		labelName := focusedOption.StringValue()
		parentIDs, err := getParentIDs(s, i.GuildID)
		if err != nil {
			log.Println("Error getting parent IDs:", err)
			return
		}
		for _, parentID := range parentIDs {
			parent, err := s.Channel(parentID)
			if err != nil {
				log.Println("Error getting channel:", err)
				continue
			}
			if strings.HasPrefix(strings.ToLower(parent.Name), strings.ToLower(labelName)) || labelName == "" {
				Choices = append(Choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  parent.Name,
					Value: parent.ID,
				})
			}
			if len(Choices) >= 25 {
				break
			}
		}
	}

	if focusedOption.Name == "name" || focusedOption.Name == "channel" {
		eventName := focusedOption.StringValue()
		channels, err := s.GuildChannels(i.GuildID)
		if err != nil {
			log.Println("Error getting channels:", err)
			return
		}
		for _, channel := range channels {
			if channel.Type != discordgo.ChannelTypeGuildText {
				continue
			}
			if strings.HasPrefix(strings.ToLower(channel.Name), strings.ToLower(eventName)) || eventName == "" {
				Choices = append(Choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  channel.Name,
					Value: channel.ID,
				})
			}
			if len(Choices) >= 25 {
				break
			}
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: Choices,
		},
	})
	if err != nil {
		log.Println("Error responding to interaction:", err)
		return
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
		eventName := message.Content[len(newEventPrefix) : len(message.Content)-len(newEventSuffix)]
		log.Println(eventName)
		role, err := getRole(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting role:", err)
			return
		}
		err = s.GuildMemberRoleAdd(r.GuildID, r.UserID, role.ID)
		if err != nil {
			log.Println("Error adding role:", err)
			return
		}
		channel, err := getChannel(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting channel:", err)
			return
		}
		message := fmt.Sprintf("<@%s> がこのイベントに参加しました！", r.UserID)
		s.ChannelMessageSend(channel.ID, message)
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
		eventName := message.Content[len(newEventPrefix) : len(message.Content)-len(newEventSuffix)]
		log.Println(eventName)
		role, err := getRole(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting role:", err)
			return
		}
		err = s.GuildMemberRoleRemove(r.GuildID, r.UserID, role.ID)
		if err != nil {
			log.Println("Error removing role:", err)
			return
		}
		channel, err := getChannel(s, r.GuildID, eventName)
		if err != nil {
			log.Println("Error getting channel:", err)
			return
		}
		message := fmt.Sprintf("<@%s> がこのイベントから離脱しました", r.UserID)
		s.ChannelMessageSend(channel.ID, message)
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

func getChannelwithInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, channelName string) (*discordgo.Channel, error) {
	// here は送信されたチャンネル
	if channelName == "here" {
		channel, err := s.Channel(i.ChannelID)
		if err != nil {
			return nil, err
		}
		return channel, nil
	}

	return getChannel(s, i.GuildID, channelName)
}

func getChannel(s *discordgo.Session, guildID string, channelName string) (*discordgo.Channel, error) {
	// もともとID形式
	channel, err := s.Channel(channelName)
	if err == nil {
		return channel, nil
	}

	// チャンネルリンク形式
	if strings.HasPrefix(channelName, "https://discord") {
		ID := strings.Split(channelName, "/")[5]
		channel, err := s.Channel(ID)
		if err != nil {
			return nil, err
		}
		return channel, nil
	}

	// チャンネル名の文字列
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}
	for _, channel := range channels {
		if channel.Name == strings.ToLower(channelName) {
			return channel, nil
		}
	}
	return nil, fmt.Errorf("channel not found")
}

func getRole(s *discordgo.Session, guildID string, roleName string) (*discordgo.Role, error) {
	// もともとID形式
	role, err := s.State.Role(guildID, roleName)
	if err == nil {
		return role, nil
	}

	// ロールメンション形式
	if strings.HasPrefix(roleName, "<@&") {
		ID := strings.TrimRight(strings.TrimLeft(roleName, "<@&"), ">")
		role, err := s.State.Role(guildID, ID)
		if err != nil {
			return nil, err
		}
		return role, nil
	}

	// ロール名の文字列
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		if role.Name == roleName {
			return role, nil
		}
	}

	return nil, fmt.Errorf("role not found")
}

func getParentfromChild(s *discordgo.Session, channelID string) (*discordgo.Channel, error) {
	channel, err := s.Channel(channelID)
	if err != nil {
		return nil, err
	}
	return getParent(s, channel.GuildID, channel.ParentID)
}

func getParentIDs(s *discordgo.Session, guildID string) ([]string, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	parentIDkeys := make(map[string]struct{})
	for _, channel := range channels {
		if channel.ParentID != "" {
			parentIDkeys[channel.ParentID] = struct{}{}
		}
	}
	var parentIDs []string
	for parentID := range parentIDkeys {
		parentIDs = append(parentIDs, parentID)
	}

	return parentIDs, nil
}

func getParent(s *discordgo.Session, guildID string, parentName string) (*discordgo.Channel, error) {
	parent, err := s.Channel(parentName)
	if err == nil {
		return parent, nil
	}

	parentIDs, err := getParentIDs(s, guildID)
	if err != nil {
		return nil, err
	}
	for _, parentID := range parentIDs {
		parent, err := s.Channel(parentID)
		if err != nil {
			return nil, err
		}
		if parent.Name == parentName {
			return parent, nil
		}
	}
	return nil, fmt.Errorf("parent channel not found")
}

func getChildIDs(s *discordgo.Session, guildID string, parentID string) ([]string, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}

	childIDs := make([]string, 0)
	for _, channel := range channels {
		if channel.ParentID == parentID {
			childIDs = append(childIDs, channel.ID)
		}
	}

	return childIDs, nil
}
