package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		return
	}
	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if err != nil {
		http.Error(w, "Failed to create secrets verifier", http.StatusBadRequest)
		return
	}
	if _, err := verifier.Write(body); err != nil {
		http.Error(w, "Failed to write to secrets verifier", http.StatusInternalServerError)
		return
	}
	if err := verifier.Ensure(); err != nil {
		http.Error(w, "Invalid request signature", http.StatusUnauthorized)
		return
	}

	event, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		http.Error(w, "Failed to parse Slack event", http.StatusBadRequest)
		return
	}

	if event.Type == slackevents.URLVerification {
		var response *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &response); err != nil {
			http.Error(w, "Failed to unmarshal URL verification event", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response.Challenge))
		return
	}

	if event.Type == slackevents.CallbackEvent {
		err = h.handleCallbackEvent(event)
		if e := new(valueobject.ErrorNotFound); errors.As(err, &e) {
			http.Error(w, e.Error(), http.StatusNotFound)
			return
		}
		if e := new(valueobject.ErrorAlreadyExists); errors.As(err, &e) {
			http.Error(w, e.Error(), http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, "Failed to handle callback event", http.StatusInternalServerError)
		}
		return
	}

	http.Error(w, "Unsupported event type", http.StatusBadRequest)
}

func (h *SlackEventHandler) handleCallbackEvent(event slackevents.EventsAPIEvent) error {
	switch e := event.InnerEvent.Data.(type) {
	case *slackevents.MessageMetadataDeletedEvent:
		if err := h.handleMessageMetadataDeletedEvent(e); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported event type: %T", e)
	}
	return nil
}

func (h *SlackEventHandler) handleMessageMetadataDeletedEvent(event *slackevents.MessageMetadataDeletedEvent) error {
	if event.PreviousMetadata.EventType != "warikan" {
		return nil
	}

	paymentIDPayload := event.PreviousMetadata.EventPayload["payment_id"]
	if paymentIDPayload == nil {
		return nil
	}
	rawPaymentID, ok := paymentIDPayload.(string)
	if !ok {
		return nil
	}

	paymentID, err := valueobject.NewPaymentIDFromString(rawPaymentID)
	if err != nil {
		return err
	}

	return h.paymentUsecase.Delete(paymentID)
}
