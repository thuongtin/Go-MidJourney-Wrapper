package main

import (
	"MJ/globals"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

type Payload struct {
	ChatbotId      string `json:"chatbotId"`
	Input          string `json:"input"`
	ConversationId string `json:"conversationId"`
	UserId         string `json:"userId"`
	Model          string `json:"model"`
	Provider       string `json:"provider"`
	IncludeHistory bool   `json:"includeHistory"`
}

type oraResponse struct {
	Response       string `json:"response"`
	ConversationID string `json:"conversationId"`
	UserID         string `json:"userId"`
	Elapsed        int    `json:"elapsed"`
}

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
	if m.Author.ID != globals.MidJourneyID {
		return
	}
	if len(m.Attachments) == 0 {
		return
	}
	spew.Dump("XXXXX", m)
	files := make([]*discordgo.File, len(m.Attachments))
	msg := m.Message.Content
	for i, attachment := range m.Attachments {
		msg += "\n" + attachment.URL
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
		Content: msg,
		//Files:      files,
		Components: modifyButtonLabels(m.ID, m.Components),
	})

	if err != nil {
		fmt.Println("Error sending message to channel:", err)
	}
}
func messageUpdateHandler(s *discordgo.Session, m *discordgo.MessageUpdate) {
	//if m.Message.Author.ID == globals.MidJourneyID && m.Author != nil {
	//	fmt.Printf("Message updated by %s: %s\n", m.Author.Username, m.Content)
	//	spew.Dump(m.Attachments)
	//}
	if m.Message.Author != nil && m.Message.Author.ID == globals.MidJourneyID {
		spew.Dump("messageUpdateHandler", m.Message)
		//spew.Dump(m.Content)
		//fmt.Printf("Message updated by %s: %s\n", m.Author.Username, m.Message.Content)
		//spew.Dump(m.Attachments)
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

func customSplit(r rune) bool {
	return r == '\n' || r == '\r'
}
