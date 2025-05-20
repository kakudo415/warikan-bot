package usecase

import (
	"fmt"

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

type Settlement struct {
	Total        valueobject.Yen
	Instructions []*SettlementInstruction
}

type SettlementInstruction struct {
	From   valueobject.PayerID
	To     valueobject.PayerID
	Amount valueobject.Yen
}

func (u *PaymentUsecase) Create(eventID valueobject.EventID, payerID valueobject.PayerID, amount valueobject.Yen) (*entity.Payment, error) {
	if eventID.IsUnknown() {
		return nil, valueobject.NewErrorNotFound("eventID is unknown", nil)
	}
	event := &entity.Event{
		ID: eventID,
	}
	if err := u.events.CreateIfNotExists(event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	if payerID.IsUnknown() {
		return nil, valueobject.NewErrorNotFound("payerID is unknown", nil)
	}
	payer := &entity.Payer{
		ID:      payerID,
		EventID: eventID,
	}
	if err := u.payers.CreateIfNotExists(payer); err != nil {
		return nil, fmt.Errorf("failed to create payer: %w", err)
	}

	payment := &entity.Payment{
		ID:      valueobject.NewPaymentID(),
		EventID: eventID,
		PayerID: payerID,
		Amount:  amount,
	}
	if err := u.payments.Create(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (u *PaymentUsecase) Delete(paymentID valueobject.PaymentID) error {
	if err := u.payments.Delete(paymentID); err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}
	return nil
}

func (u *PaymentUsecase) Join(eventID valueobject.EventID, payerID valueobject.PayerID) (*entity.Payer, error) {
	if eventID.IsUnknown() {
		return nil, valueobject.NewErrorNotFound("eventID is unknown", nil)
	}
	event := &entity.Event{
		ID: eventID,
	}
	if err := u.events.CreateIfNotExists(event); err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	if payerID.IsUnknown() {
		return nil, valueobject.NewErrorNotFound("payerID is unknown", nil)
	}
	payer := &entity.Payer{
		ID:      payerID,
		EventID: eventID,
	}
	if err := u.payers.Create(payer); err != nil {
		return nil, fmt.Errorf("failed to create payer: %w", err)
	}

	return payer, nil
}

func (u *PaymentUsecase) Settle(eventID valueobject.EventID) (*Settlement, error) {
	payments, err := u.payments.FindByEventID(eventID)
	if err != nil {
		return nil, err
	}
	payers, err := u.payers.FindByEventID(eventID)
	if err != nil {
		return nil, err
	}

	settlement := &Settlement{
		Total:        valueobject.Yen(0),
		Instructions: make([]*SettlementInstruction, 0, len(payers)),
	}

	payerMap := make(map[valueobject.PayerID]valueobject.Yen)

	for _, payment := range payments {
		settlement.Total += payment.Amount
		payerMap[payment.PayerID] += payment.Amount
	}

	for from, amount := range payerMap {
		settlement.Instructions = append(settlement.Instructions, &SettlementInstruction{
			From:   from,
			Amount: amount,
		})
	}

	return settlement, nil
}
