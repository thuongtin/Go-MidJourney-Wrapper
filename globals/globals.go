package globals

import (
	"github.com/go-resty/resty/v2"
)

const (
	DaVinciToken = "" // Your bot token
	ServerID     = ""
	SalaiToken   = "" // Your token
)

var (
	MidJourneyID = "936929561302675456" // midjourney bot id
)

func GetRestyClient() *resty.Client {
	client := resty.New()
	return client
}
