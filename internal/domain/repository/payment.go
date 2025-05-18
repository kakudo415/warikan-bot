package repository

import "github.com/kakudo415/warikan-bot/internal/domain/entity"

type PaymentRepository interface {
	CreatePayment(p *entity.Payment) error
}
