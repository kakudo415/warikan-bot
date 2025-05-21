package usecase

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kakudo415/warikan-bot/internal/domain/entity"
	"github.com/kakudo415/warikan-bot/internal/domain/valueobject"
)

type MockEventRepository struct{}

func (m *MockEventRepository) CreateIfNotExists(event *entity.Event) error {
	return nil
}

type MockPayerRepository struct {
	Payers []*entity.Payer
}

func (m *MockPayerRepository) Create(payer *entity.Payer) error {
	return nil
}

func (m *MockPayerRepository) CreateIfNotExists(payer *entity.Payer) error {
	return nil
}

func (m *MockPayerRepository) FindByEventID(eventID valueobject.EventID) ([]*entity.Payer, error) {
	return m.Payers, nil
}

type MockPaymentRepository struct {
	Payments []*entity.Payment
}

func (m *MockPaymentRepository) Create(payment *entity.Payment) error {
	return nil
}

func (m *MockPaymentRepository) Delete(paymentID valueobject.PaymentID) error {
	return nil
}

func (m *MockPaymentRepository) FindByEventID(eventID valueobject.EventID) ([]*entity.Payment, error) {
	return m.Payments, nil
}

func MustYen(amount int) valueobject.Yen {
	yen, err := valueobject.NewYen(amount)
	if err != nil {
		panic(err)
	}
	return yen
}

func TestSettle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		eventID            valueobject.EventID
		payers             []*entity.Payer
		payments           []*entity.Payment
		expectedSettlement *Settlement
	}{
		{
			name:    "OK: 1 payer, 1 payment",
			eventID: valueobject.NewEventID("event1"),
			payers: []*entity.Payer{
				{ID: valueobject.NewPayerID("payer1"), EventID: valueobject.NewEventID("event1")},
			},
			payments: []*entity.Payment{
				{ID: valueobject.NewPaymentID(), EventID: valueobject.NewEventID("event1"), Amount: MustYen(1000)},
			},
			expectedSettlement: &Settlement{
				Total:        MustYen(1000),
				Instructions: []*SettlementInstruction{},
			},
		},
		{
			name:    "OK: 2 payers, 1 payment",
			eventID: valueobject.NewEventID("event2"),
			payers: []*entity.Payer{
				{ID: valueobject.NewPayerID("payer1"), EventID: valueobject.NewEventID("event2")},
				{ID: valueobject.NewPayerID("payer2"), EventID: valueobject.NewEventID("event2")},
			},
			payments: []*entity.Payment{
				{ID: valueobject.NewPaymentID(), EventID: valueobject.NewEventID("event2"), PayerID: valueobject.NewPayerID("payer1"), Amount: MustYen(1000)},
			},
			expectedSettlement: &Settlement{
				Total: MustYen(1000),
				Instructions: []*SettlementInstruction{
					{From: valueobject.NewPayerID("payer2"), To: valueobject.NewPayerID("payer1"), Amount: MustYen(500)},
				},
			},
		},
		{
			name:    "OK: 7 payers, 3 payments",
			eventID: valueobject.NewEventID("event3"),
			payers: []*entity.Payer{
				{ID: valueobject.NewPayerID("payer1"), EventID: valueobject.NewEventID("event3")},
				{ID: valueobject.NewPayerID("payer2"), EventID: valueobject.NewEventID("event3")},
				{ID: valueobject.NewPayerID("payer3"), EventID: valueobject.NewEventID("event3")},
				{ID: valueobject.NewPayerID("payer4"), EventID: valueobject.NewEventID("event3")},
				{ID: valueobject.NewPayerID("payer5"), EventID: valueobject.NewEventID("event3")},
				{ID: valueobject.NewPayerID("payer6"), EventID: valueobject.NewEventID("event3")},
				{ID: valueobject.NewPayerID("payer7"), EventID: valueobject.NewEventID("event3")},
			},
			payments: []*entity.Payment{
				{ID: valueobject.NewPaymentID(), EventID: valueobject.NewEventID("event3"), PayerID: valueobject.NewPayerID("payer6"), Amount: MustYen(16891)},
				{ID: valueobject.NewPaymentID(), EventID: valueobject.NewEventID("event3"), PayerID: valueobject.NewPayerID("payer1"), Amount: MustYen(4332)},
				{ID: valueobject.NewPaymentID(), EventID: valueobject.NewEventID("event3"), PayerID: valueobject.NewPayerID("payer1"), Amount: MustYen(5180)},
			},
			expectedSettlement: &Settlement{
				Total: MustYen(26403),
				Instructions: []*SettlementInstruction{
					{From: valueobject.NewPayerID("payer7"), To: valueobject.NewPayerID("payer6"), Amount: MustYen(3772)},
					{From: valueobject.NewPayerID("payer5"), To: valueobject.NewPayerID("payer6"), Amount: MustYen(3772)},
					{From: valueobject.NewPayerID("payer4"), To: valueobject.NewPayerID("payer1"), Amount: MustYen(3772)},
					{From: valueobject.NewPayerID("payer3"), To: valueobject.NewPayerID("payer6"), Amount: MustYen(3772)},
					{From: valueobject.NewPayerID("payer2"), To: valueobject.NewPayerID("payer1"), Amount: MustYen(1969)},
					{From: valueobject.NewPayerID("payer2"), To: valueobject.NewPayerID("payer6"), Amount: MustYen(1803)},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			eventRepo := &MockEventRepository{}
			payerRepo := &MockPayerRepository{Payers: test.payers}
			paymentRepo := &MockPaymentRepository{Payments: test.payments}
			usecase := NewPayment(eventRepo, payerRepo, paymentRepo)

			settlement, err := usecase.Settle(test.eventID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, test.expectedSettlement.Total, settlement.Total, "total amount mismatch")
			assert.ElementsMatchf(t, test.expectedSettlement.Instructions, settlement.Instructions, "settlement instructions mismatch")
		})
	}
}
