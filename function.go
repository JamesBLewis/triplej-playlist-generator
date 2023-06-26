package triplej_playlist_generator

import (
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"github.com/JamesBLewis/triplej-playlist-generator/internal"
)

func init() {
	functions.HTTP("StartBot", StartBot)
}

// StartBot is the entry point for our cloud function
func StartBot(http.ResponseWriter, *http.Request) {
	internal.CreateBot()
}
