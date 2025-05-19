package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
	"github.com/kakudo415/warikan-bot/internal/usecase"
)

type SlackEventHandler struct {
	signingSecret  string
	client         *slack.Client
	paymentUsecase *usecase.PaymentUsecase
}

func NewSlackEventHandler(token string, signingSecret string, paymentUsecase *usecase.PaymentUsecase) *SlackEventHandler {
	return &SlackEventHandler{
		client:         slack.New(token),
		signingSecret:  signingSecret,
		paymentUsecase: paymentUsecase,
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

	if event.Type == slackevents.URLVerification {
		var response *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &response); err != nil {
			http.Error(w, "Failed to unmarshal URL verification event", http.StatusBadRequest)
			log.Println("ERROR: Failed to unmarshal URL verification event:", err)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response.Challenge))
		return
	}

	if event.Type == slackevents.CallbackEvent {
		if err := h.handleCallbackEvent(event); err != nil {
			http.Error(w, "Failed to handle callback event", http.StatusInternalServerError)
			log.Println("ERROR: Failed to handle callback event:", err)
		}
		return
	}

	http.Error(w, "Unsupported event type", http.StatusBadRequest)
	log.Println("ERROR: Unsupported event type:", event.Type)
}

func (h *SlackEventHandler) handleCallbackEvent(event slackevents.EventsAPIEvent) error {
	switch e := event.InnerEvent.Data.(type) {
	case *slackevents.MessageMetadataDeletedEvent:
		if err := h.handleMessageMetadataDeletedEvent(e); err != nil {
			log.Println("ERROR: Failed to handle message metadata deleted event:", err)
		}
	default:
		log.Println("ERROR: Unsupported event type:", e)
	}
	return nil
}

func (h *SlackEventHandler) handleMessageMetadataDeletedEvent(event *slackevents.MessageMetadataDeletedEvent) error {
	if event.PreviousMetadata.EventType != "warikan" {
		return nil
	}

	paymentIDPayload := event.PreviousMetadata.EventPayload["payment_id"]
	if paymentIDPayload == nil {
		log.Println("ERROR: payment_id not found in event payload")
		return nil
	}
	rawPaymentID, ok := paymentIDPayload.(string)
	if !ok {
		log.Println("ERROR: payment_id is not a string")
		return nil
	}

	paymentID, err := valueobject.NewPaymentIDFromString(rawPaymentID)
	if err != nil {
		log.Println("ERROR: Failed to create PaymentID from string:", err)
		return nil
	}

	if err := h.paymentUsecase.Delete(paymentID); err != nil {
		log.Println("ERROR: Failed to delete payment:", err)
		return nil
	}

	return nil
}
