package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kakudo415/warikan-bot/internal/infrastructure/handler"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/slack/events", handler.NewSlackEventHandler(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET")))
	log.Println("Starting server on 0.0.0.0:5272")
	http.ListenAndServe("0.0.0.0:5272", mux) // U+5272 = 割
}
