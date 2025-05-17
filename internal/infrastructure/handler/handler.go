package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type SlackEventHandler struct {
	client        *slack.Client
	signingSecret string
}

func NewSlackEventHandler(token string, signingSecret string) *SlackEventHandler {
	return &SlackEventHandler{
		client:        slack.New(token),
		signingSecret: signingSecret,
	}
}

func (h *SlackEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		log.Println("ERROR: Failed to read request body:", err)
		return
	}

	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if err != nil {
		http.Error(w, "Failed to create secrets verifier", http.StatusBadRequest)
		log.Println("ERROR: Failed to create secrets verifier:", err)
		return
	}
	if _, err := verifier.Write(body); err != nil {
		http.Error(w, "Failed to write to secrets verifier", http.StatusInternalServerError)
		log.Println("ERROR: Failed to write to secrets verifier:", err)
		return
	}
	if err := verifier.Ensure(); err != nil {
		http.Error(w, "Invalid request signature", http.StatusUnauthorized)
		log.Println("ERROR: Invalid request signature:", err)
		return
	}

	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		http.Error(w, "Failed to parse Slack event", http.StatusBadRequest)
		log.Println("ERROR: Failed to parse Slack event:", err)
		return
	}
	switch event.Type {
	case slackevents.URLVerification:
		var response *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &response); err != nil {
			http.Error(w, "Failed to unmarshal URL verification event", http.StatusBadRequest)
			log.Println("ERROR: Failed to unmarshal URL verification event:", err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response.Challenge))
	case slackevents.CallbackEvent:
		if err := h.HandleCallbackEvent(event); err != nil {
			http.Error(w, "Failed to handle callback event", http.StatusInternalServerError)
			log.Println("ERROR: Failed to handle callback event:", err)
			return
		}
	default:
		http.Error(w, "Unsupported event type", http.StatusBadRequest)
		log.Println("ERROR: Unsupported event type:", event.Type)
		return
	}
}

func (h *SlackEventHandler) HandleCallbackEvent(event slackevents.EventsAPIEvent) error {
	switch e := event.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		h.client.PostMessage(e.Channel, slack.MsgOptionBlocks(
			slack.NewHeaderBlock(slack.NewTextBlockObject(slack.PlainTextType, "割り勘 :thread:", true, false)),
			slack.NewSectionBlock(
				slack.NewTextBlockObject(slack.PlainTextType, "割り勘の計算を開始します。支払った金額を返信してください。", false, false),
				nil,
				nil,
			),
		))
	}
	return nil
}
