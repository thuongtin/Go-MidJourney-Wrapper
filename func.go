package main

import (
	"MJ/globals"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-resty/resty/v2"
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

func oraRequest(input string) (*resty.Response, *oraResponse, error) {
	// Khởi tạo client resty
	client := resty.New()
	data := Payload{
		ChatbotId:      "cc55ba2e-0608-4a2b-be59-8b7b645f63fd",
		Input:          input,
		ConversationId: "321966a3-511b-4895-a1ac-8337d006a2d8",
		UserId:         "95a722c9-b31c-4970-815d-03649f373b7f",
		Model:          "gpt-3.5-turbo",
		Provider:       "OPEN_AI",
		IncludeHistory: true,
	}

	jsonData, _ := json.Marshal(data)
	var oraResp oraResponse
	resp, err := client.R().
		SetHeader("authority", "ora.sh").
		SetHeader("accept", "*/*").
		SetHeader("accept-language", "en-US,en;q=0.9").
		SetHeader("content-type", "application/json").
		SetHeader("dnt", "1").
		SetHeader("origin", "https://ora.sh").
		SetHeader("referer", "https://ora.sh/juicy-olive-8tkq/coloringforkids").
		SetHeader("sec-ch-ua", `"Not:A-Brand";v="99", "Chromium";v="112"`).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-platform", "macOS").
		SetHeader("sec-fetch-dest", "empty").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-site", "same-origin").
		SetHeader("sec-gpc", "1").
		SetHeader("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36").
		SetBody(jsonData).
		SetResult(&oraResp).
		Post("https://ora.sh/api/conversation")

	// Kiểm tra lỗi
	if err != nil {
		return nil, nil, err
	}

	// Trả về kết quả
	return resp, &oraResp, nil
}

// Giải phương trình bậc 2
