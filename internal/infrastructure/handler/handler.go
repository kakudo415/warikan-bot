package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
	"github.com/kakudo415/warikan-bot/internal/usecase"
)

type SlackEventHandler struct {
	client         *slack.Client
	signingSecret  string
	paymentUsecase *usecase.PaymentUsecase
	amountPattern  *regexp.Regexp
}

func NewSlackEventHandler(token string, signingSecret string, paymentUsecase *usecase.PaymentUsecase) *SlackEventHandler {
	return &SlackEventHandler{
		client:         slack.New(token),
		signingSecret:  signingSecret,
		paymentUsecase: paymentUsecase,
		amountPattern:  regexp.MustCompile(`\b((?:\d{1,3}(?:,\d{3})+|\d+))円?\b`),
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
		return h.HandleAppMentionEvent(e)
	default:
		log.Println("ERROR : Unsupported event type:", e)
	}
	return nil
}

func (h *SlackEventHandler) HandleAppMentionEvent(event *slackevents.AppMentionEvent) error {
	eventID := valueobject.NewEventID(event.Channel)
	payerID := valueobject.NewPayerID(event.User)
	paymentID := valueobject.NewPaymentID(event.TimeStamp)

	match := h.amountPattern.FindStringSubmatch(event.Text)
	if match == nil {
		_, err := h.paymentUsecase.Join(eventID, payerID)
		if err != nil {
			log.Println("ERROR: Failed to join event:", err)
			return err
		}
		_, _, err = h.client.PostMessage(event.Channel, slack.MsgOptionText(
			"<@"+event.User+">さんが割り勘に参加しました！",
			false),
			slack.MsgOptionTS(event.TimeStamp),
		)
		if err != nil {
			log.Println("ERROR: Failed to post message to Slack:", err)
			return err
		}
	} else {
		rawAmount := strings.ReplaceAll(match[1], ",", "")
		amount, err := strconv.Atoi(rawAmount)
		if err != nil {
			log.Println("ERROR: Failed to convert amount to integer:", err)
			return err
		}
		amountYen, err := valueobject.NewYen(amount)
		if err != nil {
			log.Println("ERROR: Failed to create Yen value object:", err)
			return err
		}

		payment, err := h.paymentUsecase.Create(eventID, payerID, paymentID, amountYen)
		if err != nil {
			log.Println("ERROR: Failed to create payment:", err)
			return err
		}
		_, _, err = h.client.PostMessage(event.Channel, slack.MsgOptionText(
			"<@"+event.User+">さんの立替払い"+payment.Amount.String()+"を記録しました！",
			false),
			slack.MsgOptionTS(event.TimeStamp),
		)
		if err != nil {
			log.Println("ERROR: Failed to post message to Slack:", err)
			return err
		}
	}

	return nil
}
