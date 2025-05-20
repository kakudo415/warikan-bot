package valueobject

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type (
	EventID   struct{ value string }
	PayerID   struct{ value string }
	PaymentID struct{ value uuid.UUID }
)

func NewEventID(value string) EventID {
	return EventID{value: value}
}

func (e EventID) String() string {
	return e.value
}

func (e EventID) IsUnknown() bool {
	return e.value == ""
}

func NewPayerID(value string) PayerID {
	return PayerID{value: value}
}

func (p PayerID) String() string {
	return p.value
}

func (p PayerID) IsUnknown() bool {
	return p.value == ""
}

func NewPaymentID() PaymentID {
	return PaymentID{value: uuid.New()}
}

func NewPaymentIDFromString(value string) (PaymentID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return PaymentID{}, err
	}
	return PaymentID{value: id}, nil
}

func (p PaymentID) String() string {
	return p.value.String()
}

type Yen int64

func NewYen(amount int) (Yen, error) {
	if amount < 0 {
		return 0, errors.New("amount cannot be negative")
	}
	return Yen(amount), nil
}

func (y Yen) Int64() int64 {
	return int64(y)
}

func (y Yen) String() string {
	p := message.NewPrinter(language.Japanese)
	return p.Sprintf("%d円", number.Decimal(y.Int64()))
}

func (y Yen) MultiplyBy(multiplier int) (Yen, error) {
	if multiplier < 0 {
		return 0, errors.New("multiplier cannot be negative")
	}
	return Yen(y.Int64() * int64(multiplier)), nil
}

func (y Yen) CeilDivideBy(divisor int) (Yen, error) {
	if divisor <= 0 {
		return 0, errors.New("divisor cannot be zero or negative")
	}
	return Yen((y.Int64() + int64(divisor) - 1) / int64(divisor)), nil
}
