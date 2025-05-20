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
	if len(payers) <= 0 {
		return nil, fmt.Errorf("no payers found for eventID: %s", eventID)
	}

	settlement := &Settlement{
		Total:        valueobject.Yen(0),
		Instructions: make([]*SettlementInstruction, 0, len(payers)),
	}

	debts := make(map[valueobject.PayerID]valueobject.Yen)
	for _, payment := range payments {
		settlement.Total += payment.Amount
		debt, err := payment.Amount.CeilDivideBy(len(payers))
		if err != nil {
			return nil, fmt.Errorf("failed to divide payment amount: %w", err)
		}
		for _, payer := range payers {
			if payment.PayerID == payer.ID {
				othersDebt, err := debt.MultiplyBy(len(payers) - 1)
				if err != nil {
					return nil, fmt.Errorf("failed to multiply payment amount: %w", err)
				}
				debts[payer.ID] -= othersDebt
				continue
			}
			debts[payer.ID] += debt
		}
	}

	for {
		var maxDebterID, maxCreditorID valueobject.PayerID
		var maxDebt, maxCredit valueobject.Yen
		for payerID, debt := range debts {
			if debt >= maxDebt {
				maxDebterID = payerID
				maxDebt = debt
			}
			if debt <= maxCredit {
				maxCreditorID = payerID
				maxCredit = debt
			}
		}
		if maxDebt == 0 || maxCredit == 0 {
			break
		}

		amount := min(maxDebt, -maxCredit)

		instruction := &SettlementInstruction{
			From:   maxDebterID,
			To:     maxCreditorID,
			Amount: amount,
		}
		settlement.Instructions = append(settlement.Instructions, instruction)

		debts[maxDebterID] -= amount
		debts[maxCreditorID] += amount
	}

	return settlement, nil
}
