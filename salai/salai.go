package salai

import (
	"MJ/globals"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/go-resty/resty/v2"
)

var mux sync.Mutex

func PassPromptToSelfBot(channelID, prompt string) (*resty.Response, error) {
	mux.Lock()
	defer mux.Unlock()
	payload := map[string]interface{}{
		"type":           2,
		"application_id": "936929561302675456",
		"guild_id":       globals.ServerID,
		"channel_id":     channelID,
		"session_id":     "2fb980f65e5c9a77c96ca01f2c242cf6",
		"data": map[string]interface{}{
			"version": "1077969938624553050",
			"id":      "938956540159881230",
			"name":    "imagine",
			"type":    1,
			"options": []map[string]interface{}{
				{
					"type":  3,
					"name":  "prompt",
					"value": prompt,
				},
			},
			"application_command": map[string]interface{}{
				"id":                         "938956540159881230",
				"application_id":             "936929561302675456",
				"version":                    "1077969938624553050",
				"default_permission":         true,
				"default_member_permissions": nil,
				"type":                       1,
				"nsfw":                       false,
				"name":                       "imagine",
				"description":                "Create images with Midjourney",
				"dm_permission":              true,
				"options": []map[string]interface{}{
					{
						"type":        3,
						"name":        "prompt",
						"description": "The prompt to imagine",
						"required":    true,
					},
				},
			},
			"attachments": []interface{}{},
		},
	}

	headers := map[string]string{
		"authorization": globals.SalaiToken,
	}

	client := globals.GetRestyClient()
	response, err := client.R().
		SetHeaders(headers).
		SetBody(payload).
		Post("https://discord.com/api/v9/interactions")

	return response, err
}

func Forward(channelID, messageId, customId string) (*resty.Response, error) {
	payload := map[string]interface{}{
		"type":           3,
		"guild_id":       globals.ServerID,
		"channel_id":     channelID,
		"message_flags":  0,
		"message_id":     messageId,
		"application_id": "936929561302675456",
		"session_id":     "1f3dbdf09efdf93d81a3a6420882c92c",
		"data": map[string]interface{}{
			"component_type": 2,
			"custom_id":      customId,
		},
	}

	headers := map[string]string{
		"authorization": globals.SalaiToken,
	}

	client := globals.GetRestyClient()
	response, err := client.R().
		SetHeaders(headers).
		SetBody(payload).
		Post("https://discord.com/api/v9/interactions")

	return response, err
}

func Remix(channelID, value, customId string) (*resty.Response, error) {
	payload := map[string]interface{}{
		"type":           5,
		"guild_id":       globals.ServerID,
		"channel_id":     channelID,
		"application_id": "936929561302675456",
		"session_id":     "461f9979ffa0a1130cf93f5f42a9c725",
		"data": map[string]interface{}{
			"custom_id": customId,
			"components": []map[string]interface{}{
				{
					"type":      4,
					"custom_id": "MJ::RemixModal::new_prompt",
					"value":     value,
				},
			},
		},
	}

	headers := map[string]string{
		"authorization": globals.SalaiToken,
	}

	client := globals.GetRestyClient()
	response, err := client.R().
		SetHeaders(headers).
		SetBody(payload).
		Post("https://discord.com/api/v9/interactions")

	return response, err
}

func Describe(channelId, fileName, uploadFileName string) (*resty.Response, error) {
	url := "https://discord.com/api/v9/interactions"
	headers := map[string]string{
		"authorization": globals.SalaiToken,
		"content-type":  "multipart/form-data; boundary=----WebKitFormBoundaryTDAQP4XE7BWJ9mLB",
	}

	data := fmt.Sprintf("------WebKitFormBoundaryTDAQP4XE7BWJ9mLB\r\nContent-Disposition: form-data; name=\"payload_json\"\r\n\r\n{\"type\":2,\"application_id\":\"936929561302675456\",\"guild_id\":\"%s\",\"channel_id\":\"%s\",\"session_id\":\"6f96210dd008383ec6c17a23b6e11895\",\"data\":{\"version\":\"1092492867185950853\",\"id\":\"1092492867185950852\",\"name\":\"describe\",\"type\":1,\"options\":[{\"type\":11,\"name\":\"image\",\"value\":0}],\"application_command\":{\"id\":\"1092492867185950852\",\"application_id\":\"936929561302675456\",\"version\":\"1092492867185950853\",\"default_member_permissions\":null,\"type\":1,\"nsfw\":false,\"name\":\"describe\",\"description\":\"Writes a prompt based on your image.\",\"dm_permission\":true,\"options\":[{\"type\":11,\"name\":\"image\",\"description\":\"The image to describe\",\"required\":true}]},\"attachments\":[{\"id\":\"0\",\"filename\":\"%s\",\"uploaded_filename\":\"%s\"}]}}\r\n------WebKitFormBoundaryTDAQP4XE7BWJ9mLB--\r\n", globals.ServerID, channelId, fileName, uploadFileName)

	client := resty.New()
	resp, err := client.R().
		SetHeaders(headers).
		SetBody([]byte(data)).
		Post(url)

	return resp, err
}

func UploadToDiscord(fileName string, fileContent []byte) (string, error) {

	type UploadResp struct {
		Attachments []struct {
			ID             int    `json:"id"`
			UploadURL      string `json:"upload_url"`
			UploadFilename string `json:"upload_filename"`
		} `json:"attachments"`
	}
	client := globals.GetRestyClient()
	headers := map[string]string{
		"authorization": globals.SalaiToken,
	}

	var uploadResp UploadResp
	// Make the request to upload the file
	_, err := client.R().
		//SetBody(buf.Bytes()).
		SetHeaders(headers).
		SetBody(map[string]interface{}{
			"files": []map[string]interface{}{
				{
					"filename":  fileName,
					"file_size": len(fileContent),
					"id":        "10",
				},
			},
		}).SetResult(&uploadResp).
		Post("https://discord.com/api/v9/channels/1093452634356203530/attachments")
	if err != nil {
		return "", err
	}
	if len(uploadResp.Attachments) == 0 {
		return "", errors.New("no attachments found")
	}
	_, err = client.R().
		//SetBody(buf.Bytes()).
		//SetHeaders(headers).
		SetBody(fileContent).
		Put(uploadResp.Attachments[0].UploadURL)
	if err != nil {
		log.Panicln(err)
		return "", err
	}
	//uploadResp.Attachments[0].UploadURL

	return uploadResp.Attachments[0].UploadFilename, nil
}
