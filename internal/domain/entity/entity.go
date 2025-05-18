package entity

import (
	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
)

type Event struct {
	ID valueobject.EventID
}

func NewEvent() *Event {
	return &Event{
		ID: valueobject.NewEventID(),
	}
}

type Payer struct {
	ID valueobject.PayerID
}

func NewPayer() *Payer {
	return &Payer{
		ID: valueobject.NewPayerID(),
	}
}

type Payment struct {
	ID      valueobject.PaymentID
	EventID valueobject.EventID
	PayerID valueobject.PayerID
	Amount  valueobject.Yen
}

func NewPayment(eventID valueobject.EventID, payerID valueobject.PayerID, amount valueobject.Yen) *Payment {
	return &Payment{
		ID:      valueobject.NewPaymentID(),
		EventID: eventID,
		PayerID: payerID,
		Amount:  amount,
	}
}
