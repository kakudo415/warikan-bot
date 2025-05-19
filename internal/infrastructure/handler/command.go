package handler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/slack-go/slack"

	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
	"github.com/kakudo415/warikan-bot/internal/usecase"
)

type SlackCommandHandler struct {
	signingSecret  string
	client         *slack.Client
	paymentUsecase *usecase.PaymentUsecase
	amountPattern  *regexp.Regexp
}

func NewSlackCommandHandler(token string, signingSecret string, paymentUsecase *usecase.PaymentUsecase) *SlackCommandHandler {
	return &SlackCommandHandler{
		client:         slack.New(token),
		signingSecret:  signingSecret,
		paymentUsecase: paymentUsecase,
		amountPattern:  regexp.MustCompile(`\b((?:\d{1,3}(?:,\d{3})+|\d+))円?\b`),
	}
}

func (h *SlackCommandHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if err != nil {
		log.Println("ERROR: Failed to create secrets verifier: ", err)
		http.Error(w, "Failed to create secrets verifier", http.StatusBadRequest)
		return
	}
	r.Body = io.NopCloser(io.TeeReader(r.Body, &verifier))
	slash, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Println("ERROR: Failed to parse slash command: ", err)
		http.Error(w, "Failed to parse slash command", http.StatusBadRequest)
		return
	}
	if verifier.Ensure() != nil {
		log.Println("ERROR: Invalid request signature: ", err)
		http.Error(w, "Invalid request signature", http.StatusUnauthorized)
		return
	}

	err = h.HandleSlashCommand(slash)
	if err != nil {
		log.Println("ERROR: Failed to handle slash command: ", err)
		http.Error(w, "Failed to handle slash command", http.StatusInternalServerError)
		return
	}
}

func (h *SlackCommandHandler) HandleSlashCommand(slash slack.SlashCommand) error {
	switch slash.Command {
	case "/warikan":
		return h.HandleWarikanCommand(slash)
	default:
		log.Println("ERROR: Unsupported command: ", slash.Command)
		return fmt.Errorf("unsupported command: %s", slash.Command)
	}
}

func (h *SlackCommandHandler) HandleWarikanCommand(slash slack.SlashCommand) error {
	eventID := valueobject.NewEventID(slash.ChannelID)
	payerID := valueobject.NewPayerID(slash.UserID)

	match := h.amountPattern.FindStringSubmatch(slash.Text)
	if match != nil {
		rawAmount := strings.ReplaceAll(match[1], ",", "")
		amount, err := strconv.Atoi(rawAmount)
		if err != nil {
			log.Println("ERROR: Failed to convert amount to integer: ", err)
			return err
		}
		amountYen, err := valueobject.NewYen(amount)
		if err != nil {
			log.Println("ERROR: Failed to create Yen value object: ", err)
			return err
		}

		payment, err := h.paymentUsecase.Create(eventID, payerID, amountYen)
		if err != nil {
			log.Println("ERROR: Failed to create payment: ", err)
			return err
		}

		_, _, err = h.client.PostMessage(slash.ChannelID, slack.MsgOptionBlocks(
			slack.NewSectionBlock(
				slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("<@%s>さんが%s立て替えました！", slash.UserID, amountYen.String()), false, false),
				nil,
				nil,
			),
		), slack.MsgOptionMetadata(slack.SlackMetadata{
			EventType: "warikan",
			EventPayload: map[string]any{
				"payment_id": payment.ID.String(),
			},
		}))
		if err != nil {
			log.Println("ERROR: Failed to post message: ", err)
			return err
		}

		return nil
	}

	return errors.New("Invalid argument")
}
