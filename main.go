package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kakudo415/warikan-bot/internal/infrastructure/handler"
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
	slackCommandHandler := handler.NewSlackCommandHandler(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET"), paymentUsecase)
	slackEventHandler := handler.NewSlackEventHandler(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_SIGNING_SECRET"), paymentUsecase)

	mux := http.NewServeMux()
	mux.Handle("/slack/command", slackCommandHandler)
	mux.Handle("/slack/event", slackEventHandler)
	log.Println("Starting server on 0.0.0.0:5272")
	if err := http.ListenAndServe("0.0.0.0:5272", mux); err != nil { // U+5272 = 割
		log.Fatalf("server failed to start: %v", err)
	}
}
