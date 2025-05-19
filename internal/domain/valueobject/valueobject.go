package valueobject

import (
	"errors"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

type (
	EventID   struct{ value string }
	PayerID   struct{ value string }
	PaymentID struct{ value string }
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

func NewPaymentID(value string) PaymentID {
	return PaymentID{value: value}
}

func (p PaymentID) String() string {
	return p.value
}

func (p PaymentID) IsUnknown() bool {
	return p.value == ""
}

type Yen uint64

func NewYen(amount int) (Yen, error) {
	if amount < 0 {
		return 0, errors.New("amount cannot be negative")
	}
	return Yen(uint64(amount)), nil
}

func (y Yen) Uint64() uint64 {
	return uint64(y)
}

func (y Yen) String() string {
	p := message.NewPrinter(language.Japanese)
	return p.Sprintf("%d円", number.Decimal(y.Uint64()))
}
