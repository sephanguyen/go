package commands

type CancelReserveClassCommandPayload struct {
	StudentPackageID string
	StudentID        string
	CourseID         string
	ActiveClassID    string
}
