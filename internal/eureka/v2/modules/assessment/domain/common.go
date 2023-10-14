package domain

type Error string

// Error ToString method
func (e Error) Error() string {
	return string(e)
}

// Error codes.
const (
	ErrIDRequired                  Error = "ID required"
	ErrCourseIDRequired            Error = "course ID required"
	ErrLearningMaterialIDRequired  Error = "learning material ID required"
	ErrInvalidLearningMaterialType Error = "invalid learning material type"
	ErrInvalidGradingStatus        Error = "invalid grading status"
	ErrSubmissionIDRequired        Error = "submission ID required"
	ErrAssessmentIDRequired        Error = "assessment ID required"
	ErrStudentIDRequired           Error = "student ID required"
	ErrUserIDRequired              Error = "user ID required"
	ErrInvalidSessionStatus        Error = "invalid session status"
	ErrCompletedAtRequired         Error = "completed at required"
)
