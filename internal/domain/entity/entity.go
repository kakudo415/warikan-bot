package entity

import (
	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
)

type Event struct {
	ID valueobject.EventID
}

type Payer struct {
	ID      valueobject.PayerID
	EventID valueobject.EventID
}

type Payment struct {
	ID      valueobject.PaymentID
	EventID valueobject.EventID
	PayerID valueobject.PayerID
	Amount  valueobject.Yen
}
