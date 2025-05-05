package webexcrawler

import (
	"net/http"
	"os"
)

type Crawler struct {
	ApiKey string
	client http.Client
}

func NewCrawler() *Crawler {
	apikey := os.Getenv("WEBEX_APIKEY")

	return &Crawler{
		ApiKey: apikey,
		client: *http.DefaultClient,
	}
}
