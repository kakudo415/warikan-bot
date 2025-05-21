package valueobject

import "github.com/google/uuid"

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
