package main

import (
	"MJ/globals"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func downloadAttachment(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return fileData, nil
}

func modifyButtonLabels(messageId string, components []discordgo.MessageComponent) []discordgo.MessageComponent {
	modifiedComponents := make([]discordgo.MessageComponent, len(components))

	for i, component := range components {
		switch ct := component.Type(); {
		case discordgo.ActionsRowComponent == ct:
			modifiedComponents[i] = &discordgo.ActionsRow{
				Components: modifyButtonLabels(messageId, component.(*discordgo.ActionsRow).Components),
			}
		case discordgo.ButtonComponent == ct:
			modifiedButton := *component.(*discordgo.Button)
			modifiedButton.Label = "MJ: " + modifiedButton.Label
			if modifiedButton.CustomID != "" {
				modifiedButton.CustomID = messageId + "_" + modifiedButton.CustomID
			}
			modifiedComponents[i] = modifiedButton
		default:
			modifiedComponents[i] = component
		}
	}

	return modifiedComponents
}

func messageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == "" {
		return
	}
	if strings.HasPrefix(m.Content, "$mj_target") && m.MessageReference != nil {
		targetMessage, err := s.ChannelMessage(m.MessageReference.ChannelID, m.MessageReference.MessageID)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Exception has occured, maybe you didn't reply to MidJourney message")
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}
		if targetMessage.Author.ID != globals.MidJourneyID {
			s.ChannelMessageSend(m.ChannelID, "Use the command only when you reply to MidJourney")
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}

		s.ChannelMessageSend(m.ChannelID, "Done")
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	} else {
		if m.Author.ID != globals.MidJourneyID {
			return
		}

		// Check if the message has attachments
		if len(m.Attachments) == 0 {
			return
		}
		//spew.Dump("MJ", m)
		// Iterate through users map

		files := make([]*discordgo.File, len(m.Attachments))
		for i, attachment := range m.Attachments {
			fileData, err := downloadAttachment(attachment.URL)
			if err != nil {
				fmt.Println("Error downloading attachment:", err)
				return
			}

			files[i] = &discordgo.File{
				Name:        attachment.Filename,
				ContentType: "image/png",
				Reader:      bytes.NewReader(fileData),
			}
		}

		_, err := s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:    m.Message.Content,
			Files:      files,
			Components: modifyButtonLabels(m.ID, m.Components),
		})
		if err != nil {
			fmt.Println("Error sending message to channel:", err)
		}
	}
}

func modifyButtonStyle(messageID string, buttonData discordgo.MessageComponentInteractionData, components []discordgo.MessageComponent, style discordgo.ButtonStyle) []discordgo.MessageComponent {
	modifiedComponents := make([]discordgo.MessageComponent, len(components))
	for i, component := range components {
		switch ct := component.Type(); {
		case discordgo.ActionsRowComponent == ct:
			modifiedComponents[i] = &discordgo.ActionsRow{
				Components: modifyButtonStyle(messageID, buttonData, component.(*discordgo.ActionsRow).Components, style),
			}
		case discordgo.ButtonComponent == ct:
			button := component.(*discordgo.Button)
			if strings.HasPrefix(button.CustomID, messageID+"_") && button.CustomID == buttonData.CustomID {
				button.Style = style
			}
			modifiedComponents[i] = button
		default:
			modifiedComponents[i] = component
		}
	}

	return modifiedComponents
}
