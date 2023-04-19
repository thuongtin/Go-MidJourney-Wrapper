package globals

import (
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

var (
	DaVinciToken, ServerID, SalaiToken string
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	DaVinciToken = os.Getenv("DISCORD_BOT_TOKEN")
	ServerID = os.Getenv("DISCORD_SERVER_ID")
	SalaiToken = os.Getenv("DISCORD_USER_TOKEN")

}

var (
	MidJourneyID = "936929561302675456" // midjourney bot id
)

func GetRestyClient() *resty.Client {
	client := resty.New()
	return client
}
