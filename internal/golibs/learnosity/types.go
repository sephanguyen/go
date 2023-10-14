package learnosity

// RenderingType dictates whether the activity will be rendered using the Learnosity's assessment player,
// or Items embedded in multiple different locations.
type RenderingType string

const (
	RenderingTypeAssess RenderingType = "assess"
	RenderingTypeInline RenderingType = "inline"
)

// SessionStatus represents valid statuses (Incomplete, Completed, Discarded, PendingScoring).
type SessionStatus string

const (
	SessionStatusNone           SessionStatus = ""
	SessionStatusIncomplete     SessionStatus = "Incomplete"
	SessionStatusCompleted      SessionStatus = "Completed"
	SessionStatusDiscarded      SessionStatus = "Discarded"
	SessionStatusPendingScoring SessionStatus = "PendingScoring"
)
