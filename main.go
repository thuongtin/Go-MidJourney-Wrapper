package main

import (
	"MJ/globals"
	"MJ/salai"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/davecgh/go-spew/spew"
)

func buttonClickedHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	buttonData := i.MessageComponentData()
	//if strings.Contains(buttonData.CustomID, "::variation::") {
	//	displayDiscordDialog(buttonData, s, i)
	//	return
	//}

	messageId := strings.Split(buttonData.CustomID, "_")[0]
	customId := strings.TrimPrefix(buttonData.CustomID, messageId+"_")
	r, err := salai.Forward(i.ChannelID, messageId, customId)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(r)
	}

	// Thay đổi nút người dùng đã click bằng màu xanh lá cây
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Components: modifyButtonStyle(messageId, buttonData, i.Message.Components, discordgo.SuccessButton),
		},
	})
}

func handleTextInputSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}
	data := i.Data.(discordgo.ModalSubmitInteractionData)

	customId := data.CustomID
	spew.Dump(salai.Remix(i.ChannelID, data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value, customId))

	//Just close the modal
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: 1 << 6,
		},
	})

}

func displayDiscordDialog(buttonData discordgo.MessageComponentInteractionData, s *discordgo.Session, i *discordgo.InteractionCreate) {
	buttonCustomId := buttonData.CustomID //"MJ::JOB::variation::2::efcff9bf-42de-42a7-ac41-668ca5d19db4"
	split := strings.Split(buttonCustomId, "::")
	//Get the last 2 elements
	customId := split[len(split)-1] + "::" + split[len(split)-2]
	//Create the new customId
	newCustomId := "MJ::RemixModal::" + customId
	fmt.Println(newCustomId)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			Title:    "Remix prompt",
			CustomID: newCustomId,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "MJ::RemixModal::new_prompt",
							Label:       "New prompt for the image",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "Enter a new prompt",
							Value:       strings.Split(strings.Split(i.Message.Content, "**")[1], "**")[0],
							Required:    true,
							MaxLength:   4000,
						},
					},
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func mjImagineCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	prompt := i.ApplicationCommandData().Options[0].StringValue()
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("```%s```\nYour image is being prepared, please wait a moment...", prompt),
		},
	})
	response, err := salai.PassPromptToSelfBot(i.ChannelID, prompt)
	if err != nil || response.StatusCode() >= 400 {
		spew.Dump(err, response)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Request has failed; please try later",
			},
		})
		return
	}
}

// Define other command handler functions here (mj_upscale_to_max, mj_variation)

func commandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	switch i.ApplicationCommandData().Name {
	case "mj_imagine":
		mjImagineCommand(s, i)
	case "mj_describe":
		mjDescribeCommand(s, i)
	//case "cc":
	//	mjColoringForChildrenCommand(s, i)
	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command",
			},
		})
	}
}

func mjDescribeCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Println("Error responding to interaction:", err)
		return
	}

	// Process the mj_describe command
	// You can access the options using `i.ApplicationCommandData().Options`
	// For example, to get the "image" option value:
	var imageID string
	for _, option := range i.ApplicationCommandData().Options {
		if option.Name == "image" {
			imageID = option.Value.(string)
			break
		}
	}

	// Do something with the imageID, like getting the attachment:
	attachment, ok := i.ApplicationCommandData().Resolved.Attachments[imageID]
	if ok {
		imgContent, err := downloadAttachment(fmt.Sprintf("%s", attachment.URL))
		if err == nil {
			fn, err := salai.UploadToDiscord(attachment.Filename, imgContent)
			if err == nil {
				salai.Describe(i.ChannelID, attachment.Filename, fn)
			}
		}
	} else {
		fmt.Println("Attachment not found.")
	}
	str := "Ok"
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &str,
	})
	if err != nil {
		panic(err)
	}
}
func main() {
	dg, err := discordgo.New("Bot " + globals.DaVinciToken)

	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "hello",
			Description: "Say hello",
		},
		{
			Name:        "mj_imagine",
			Description: "Create an image from a text prompt",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "prompt",
					Description: "This command is a wrapper of MidJourneyAI",
					Required:    true,
				},
			},
		},
		//{
		//	Name:        "cc",
		//	Description: "Create a coloring page for children",
		//	Options: []*discordgo.ApplicationCommandOption{
		//		{
		//			Type:        discordgo.ApplicationCommandOptionString,
		//			Name:        "keywords",
		//			Description: "This command will automatically generate a coloring page for children with the keywords entered",
		//			Required:    true,
		//		},
		//	},
		//},
		{
			Name:        "mj_describe",
			Description: "Create an image from a text prompt",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "image",
					Description: "Please attach an image",
					Required:    true,
				},
			},
		},
		// Add other commands here (mj_upscale_to_max, mj_variation, ...)
	}
	//dg.ApplicationCommandCreate()

	dg.AddHandler(commandHandler)
	dg.AddHandler(messageCreateHandler)
	dg.AddHandler(messageUpdateHandler)
	dg.AddHandler(buttonClickedHandler)
	dg.AddHandler(handleTextInputSubmit)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	dg.ApplicationCommandBulkOverwrite(dg.State.User.ID, globals.ServerID, commands)

	fmt.Println("Bot is now running. Press CTRL+C to exit.")
	defer dg.Close()

	<-make(chan struct{})
}
