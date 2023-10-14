package domain

import "time"

type StudentEnrollmentStatusHistory struct {
	StudentID        string
	LocationID       string
	EnrollmentStatus string
	StartDate        time.Time
	EndDate          time.Time
	Comment          string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type StudentEnrollmentStatusHistories []*StudentEnrollmentStatusHistory

func (s StudentEnrollmentStatusHistories) GetStudentEnrollmentMap() map[string][]*StudentEnrollmentStatusHistory {
	studentEnrollmentMap := make(map[string][]*StudentEnrollmentStatusHistory, len(s))

	for _, studentESH := range s {
		studentEnrollmentMap[studentESH.StudentID] = append(studentEnrollmentMap[studentESH.StudentID], studentESH)
	}

	return studentEnrollmentMap
}
