package errcode

const (
	DomainCodeOK       int = 20000
	DomainCodeInvalid  int = 40000
	DomainCodeNotFound int = 40400
	DomainCodeInternal int = 50000
)

type DomainError interface {
	DomainError() string
	DomainCode() int
}

type UnHandledError struct {
}
