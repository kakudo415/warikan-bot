package usecase

import (
	"github.com/kakudo415/warikan-bot/internal/domain/entity"
	"github.com/kakudo415/warikan-bot/internal/domain/repository"
	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
)

type PaymentUsecase struct {
	events   repository.EventRepository
	payers   repository.PayerRepository
	payments repository.PaymentRepository
}

func NewPayment(events repository.EventRepository, payers repository.PayerRepository, payments repository.PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{
		events,
		payers,
		payments,
	}
}

func (u *PaymentUsecase) Create(eventID valueobject.EventID, payerID valueobject.PayerID, amount valueobject.Yen) (*entity.Payment, error) {
	var event *entity.Event
	if eventID == valueobject.EventIDUnknown {
		event = entity.NewEvent()
		err := u.events.CreateEvent(event)
		if err != nil {
			return nil, err
		}
		eventID = event.ID
	}
	var payer *entity.Payer
	if payerID == valueobject.PayerIDUnknown {
		payer = entity.NewPayer()
		err := u.payers.CreatePayer(payer)
		if err != nil {
			return nil, err
		}
		payerID = payer.ID
	}
	payment := entity.NewPayment(eventID, payerID, amount)
	err := u.payments.CreatePayment(payment)
	if err != nil {
		return nil, err
	}
	return payment, nil
}
