package valueobject

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type (
	EventID   uuid.UUID
	PayerID   uuid.UUID
	PaymentID uuid.UUID
)

var (
	EventIDUnknown   = EventID(uuid.Nil)
	PayerIDUnknown   = PayerID(uuid.Nil)
	PaymentIDUnknown = PaymentID(uuid.Nil)
)

func NewEventID() EventID {
	return EventID(uuid.New())
}

func (e EventID) String() string {
	return uuid.UUID(e).String()
}

func EventIDFromString(id string) (EventID, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return EventIDUnknown, err
	}
	return EventID(uuid), nil
}

func NewPayerID() PayerID {
	return PayerID(uuid.New())
}

func (p PayerID) String() string {
	return uuid.UUID(p).String()
}

func PayerIDFromString(id string) (PayerID, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return PayerIDUnknown, err
	}
	return PayerID(uuid), nil
}

func NewPaymentID() PaymentID {
	return PaymentID(uuid.New())
}

func (p PaymentID) String() string {
	return uuid.UUID(p).String()
}

func PaymentIDFromString(id string) (PaymentID, error) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		return PaymentIDUnknown, err
	}
	return PaymentID(uuid), nil
}

type Yen struct {
	amount uint64
}

func NewYen(amount int) (Yen, error) {
	if amount < 0 {
		return Yen{}, errors.New("amount cannot be negative")
	}
	return Yen{uint64(amount)}, nil
}

func (y Yen) Uint64() uint64 {
	return y.amount
}

func (y Yen) String() string {
	p := message.NewPrinter(language.Japanese)
	return p.Sprintf("%d円", number.Decimal(y.amount))
}
