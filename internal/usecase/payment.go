package usecase

import (
	"errors"
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
		fmt.Println("ERROR: eventID is unknown")
		return nil, errors.New("eventID is unknown")
	}
	event := &entity.Event{
		ID: eventID,
	}
	if err := u.events.CreateIfNotExists(event); err != nil {
		fmt.Println("ERROR: Failed to create event:", err)
		return nil, err
	}

	if payerID.IsUnknown() {
		fmt.Println("ERROR: payerID is unknown")
		return nil, errors.New("payerID is unknown")
	}
	payer := &entity.Payer{
		ID:      payerID,
		EventID: eventID,
	}
	if err := u.payers.CreateIfNotExists(payer); err != nil {
		fmt.Println("ERROR: Failed to create payer:", err)
		return nil, err
	}

	payment := &entity.Payment{
		ID:      valueobject.NewPaymentID(),
		EventID: eventID,
		PayerID: payerID,
		Amount:  amount,
	}
	if err := u.payments.Create(payment); err != nil {
		fmt.Println("ERROR: Failed to create payment:", err)
		return nil, err
	}

	return payment, nil
}

func (u *PaymentUsecase) Delete(paymentID valueobject.PaymentID) error {
	if err := u.payments.Delete(paymentID); err != nil {
		fmt.Println("ERROR: Failed to delete payment:", err)
		return err
	}
	return nil
}

func (u *PaymentUsecase) Join(eventID valueobject.EventID, payerID valueobject.PayerID) (*entity.Payer, error) {
	if eventID.IsUnknown() {
		fmt.Println("ERROR: eventID is unknown")
		return nil, errors.New("eventID is unknown")
	}
	event := &entity.Event{
		ID: eventID,
	}
	u.events.CreateIfNotExists(event)

	if payerID.IsUnknown() {
		fmt.Println("ERROR: payerID is unknown")
		return nil, errors.New("payerID is unknown")
	}
	payer := &entity.Payer{
		ID:      payerID,
		EventID: eventID,
	}
	err := u.payers.Create(payer)
	if err != nil {
		fmt.Println("ERROR: Failed to create payer:", err)
		return nil, err
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
