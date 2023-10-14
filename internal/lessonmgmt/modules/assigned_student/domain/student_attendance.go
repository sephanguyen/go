package domain

import "time"

type Filter struct {
	StudentID    []string
	CourseID     []string
	LocationID   []string
	AttendStatus []string
	StartDate    time.Time
	EndDate      time.Time
}

type StudentAttendance struct {
	StudentID           string
	LessonID            string
	AttendStatus        string
	CourseID            string
	LocationID          string
	ReallocatedLessonID string
}

func (s *StudentAttendance) GetKey() string {
	return s.LessonID + "-" + s.StudentID
}

func (s *StudentAttendance) SetReallocatedLessonID(reallocatedMap map[string]string) {
	s.ReallocatedLessonID = reallocatedMap[s.GetKey()]
}

type GetStudentAttendanceParams struct {
	Limit                     int
	Offset                    int
	SearchKey                 string
	Timezone                  string
	CourseID                  []string
	LocationID                []string
	StudentID                 []string
	AttendStatus              []string
	StartDate                 time.Time
	EndDate                   time.Time

	// academic year filter
	IsFilterByCurrentYear bool
	YearStartDate         time.Time
	YearEndDate           time.Time
}
