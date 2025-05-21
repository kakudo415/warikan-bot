package valueobject

type ErrorNotFound struct {
	message string
	err     error
}

func NewErrorNotFound(message string, err error) *ErrorNotFound {
	return &ErrorNotFound{message, err}
}

func (e *ErrorNotFound) Error() string {
	if e.err != nil {
		return e.message + " (" + e.err.Error() + ")"
	}
	return e.message
}

func (e *ErrorNotFound) Unwrap() error {
	return e.err
}

type ErrorAlreadyExists struct {
	message string
	err     error
}

func NewErrorAlreadyExists(message string, err error) *ErrorAlreadyExists {
	return &ErrorAlreadyExists{message, err}
}

func (e *ErrorAlreadyExists) Error() string {
	if e.err != nil {
		return e.message + " (" + e.err.Error() + ")"
	}
	return e.message
}

func (e *ErrorAlreadyExists) Unwrap() error {
	return e.err
}

type ErrorInvalid struct {
	message string
	err     error
}
