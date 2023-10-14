package learnosity

// Error is a SDK error encountered.
type Error string

// General errors.
const (
	ErrJSONMarshalToString Error = "JSONMarshalToString: %w"
	ErrNotFoundEndpoint    Error = "the endpoint wasn't found"
)

// Error implements the error interface.
func (e Error) Error() string {
	return string(e)
}

// HTTPCode represents a status codes indicate whether a specific HTTP request has been successfully completed.
type HTTPCode uint32

const (
	HTTPTooManyRequests HTTPCode = 42000
)
