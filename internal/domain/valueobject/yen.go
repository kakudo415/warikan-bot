package valueobject

import (
	"errors"
	"fmt"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

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
		return 0, fmt.Errorf("divisor cannot be zero or negative (%d / %d)", y.Int64(), divisor)
	}
	return Yen((y.Int64() + int64(divisor) - 1) / int64(divisor)), nil
}
