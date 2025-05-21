package repository

import (
	"github.com/kakudo415/warikan-bot/internal/domain/entity"
	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
)

type EventRepository interface {
	CreateIfNotExists(event *entity.Event) error
}

type PayerRepository interface {
	Create(payer *entity.Payer) error
	CreateIfNotExists(payer *entity.Payer) error
	FindByEventID(eventID valueobject.EventID) ([]*entity.Payer, error)
}

type PaymentRepository interface {
	Create(payment *entity.Payment) error
	Delete(paymentID valueobject.PaymentID) error
	FindByEventID(eventID valueobject.EventID) ([]*entity.Payment, error)
}
