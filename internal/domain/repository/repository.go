package repository

import "github.com/kakudo415/warikan-bot/internal/domain/entity"

type EventRepository interface {
	CreateEvent(e *entity.Event) error
}

type PayerRepository interface {
	CreatePayer(p *entity.Payer) error
}

type PaymentRepository interface {
	CreatePayment(p *entity.Payment) error
}
