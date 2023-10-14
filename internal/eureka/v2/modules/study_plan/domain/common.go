package domain

type Error string

// Error ToString method
func (e Error) Error() string {
	return string(e)
}

// Error codes
const (
	ErrIDRequired        Error = "ID required"
	ErrCourseIDRequired  Error = "course ID required"
	ErrStudentIDRequired Error = "student ID required"
	ErrUserIDRequired    Error = "user ID required"
)
