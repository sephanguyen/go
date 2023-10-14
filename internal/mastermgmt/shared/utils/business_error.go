package utils

const (
	SystemError = "SystemError"
)

type BusinessError struct {
	Name  string
	Error error
}

func (b *BusinessError) Is(errType string) bool {
	return b.Name == errType
}

func NewError(name string, err error) *BusinessError {
	return &BusinessError{
		Name:  name,
		Error: err,
	}
}

func NewSystemError(err error) *BusinessError {
	return &BusinessError{
		Name:  SystemError,
		Error: err,
	}
}
