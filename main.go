package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kakudo415/warikan-bot/internal/infrastructure/handler"
	"github.com/kakudo415/warikan-bot/internal/infrastructure/mapper"
	"github.com/kakudo415/warikan-bot/internal/infrastructure/repository"
	"github.com/kakudo415/warikan-bot/internal/usecase"
)

func main() {
	eventRepository, err := repository.NewEventRepository("database.db")
	if err != nil {
		log.Fatalf("failed to create event repository: %v", err)
	}
	payerRepository, err := repository.NewPayerRepository("database.db")
	if err != nil {
		log.Fatalf("failed to create payer repository: %v", err)
	}
	paymentRepository, err := repository.NewPaymentRepository("database.db")
	if err != nil {
		log.Fatalf("failed to create payment repository: %v", err)
	}
	paymentUsecase := usecase.NewPayment(eventRepository, payerRepository, paymentRepository)
	slackMapper, err := mapper.NewSlackMapper("database.db")
	if err != nil {
		log.Fatalf("failed to create slack mapper: %v", err)
	}
	slackEventHandler := handler.NewSlackEventHandler(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET"), paymentUsecase, slackMapper)

	mux := http.NewServeMux()
	mux.Handle("/slack/events", slackEventHandler)
	log.Println("Starting server on 0.0.0.0:5272")
	http.ListenAndServe("0.0.0.0:5272", mux) // U+5272 = 割
}
