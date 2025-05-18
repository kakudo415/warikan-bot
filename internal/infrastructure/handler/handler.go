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
	"github.com/kakudo415/warikan-bot/internal/infrastructure/mapper"
	"github.com/kakudo415/warikan-bot/internal/usecase"
)

type SlackEventHandler struct {
	client         *slack.Client
	signingSecret  string
	paymentUsecase *usecase.PaymentUsecase
	slackMapper    *mapper.SlackMapper
	amountPattern  *regexp.Regexp
}

func NewSlackEventHandler(token string, signingSecret string, paymentUsecase *usecase.PaymentUsecase, slackMapper *mapper.SlackMapper) *SlackEventHandler {
	return &SlackEventHandler{
		client:         slack.New(token),
		signingSecret:  signingSecret,
		paymentUsecase: paymentUsecase,
		slackMapper:    slackMapper,
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
	match := h.amountPattern.FindStringSubmatch(event.Text)
	if match == nil {
		log.Println("INFO: No amount found")
		return nil
	}
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
	eventID, err := h.slackMapper.SlackChannelIDToEventID(event.Channel)
	if err != nil {
		log.Println("ERROR: Failed to get event ID from Slack channel ID:", err)
		return err
	}
	payerID, err := h.slackMapper.SlackUserIDToPayerID(event.User)
	if err != nil {
		log.Println("ERROR: Failed to get payer ID from Slack user ID:", err)
		return err
	}
	payment, err := h.paymentUsecase.Create(eventID, payerID, amountYen)
	if err != nil {
		log.Println("ERROR: Failed to create payment:", err)
		return err
	}
	if eventID == valueobject.EventIDUnknown {
		err := h.slackMapper.CreateEventIDSlackChannelIDMapping(payment.EventID, event.Channel)
		if err != nil {
			log.Println("ERROR: Failed to create event ID to Slack channel ID mapping:", err)
		}
	}
	if payerID == valueobject.PayerIDUnknown {
		err := h.slackMapper.CreatePayerIDSlackUserIDMapping(payment.PayerID, event.User)
		if err != nil {
			log.Println("ERROR: Failed to create payer ID to Slack user ID mapping:", err)
		}
	}
	err = h.slackMapper.CreatePaymentIDSlackThreadTSMapping(payment.ID, event.TimeStamp)
	if err != nil {
		log.Println("ERROR: Failed to create payment ID to Slack thread timestamp mapping:", err)
	}
	_, _, err = h.client.PostMessage(event.Channel, slack.MsgOptionText(
		"<@"+event.User+">さんの立替払い"+strconv.Itoa(int(payment.Amount.Uint64()))+"円を記録しました！",
		false),
		slack.MsgOptionTS(event.TimeStamp),
	)
	if err != nil {
		log.Println("ERROR: Failed to post message to Slack:", err)
		return err
	}
	return nil
}
