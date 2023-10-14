package domain

import "time"

type EnrolledStudent struct {
	StudentID        string
	EnrollmentStatus string
	CourseID         string
	LocationID       string
	StartAt          time.Time
	EndAt            time.Time
}
