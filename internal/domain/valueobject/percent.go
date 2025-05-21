package valueobject

import "errors"

type Percent uint

func NewPercent(value int) (Percent, error) {
	if value < 0 {
		return 0, errors.New("percent cannot be negative")
	}
	return Percent(value), nil
}

func (p Percent) Int() int {
	return int(p)
}
