package domain

import "time"

type ReallocateStudent struct {
	StudentId        string
	OriginalLessonID string
	CourseID         string
	LocationID       string
	ClassID          string
	GradeID          string
	StartAt          time.Time
	EndAt            time.Time
}

type RetrieveStudentPendingReallocateDto struct {
	Limit                     int
	Offset                    int
	LessonDate                time.Time
	SearchKey                 string
	Timezone                  string
	CourseID                  []string
	LocationID                []string
	GradeID                   []string
	ClassID                   []string
	StartDate                 time.Time
	EndDate                   time.Time
}

type Filter struct {
	CourseID   []string
	LocationID []string
	ClassId    []string
	GradeID    []string
	StartDate  time.Time
	EndDate    time.Time
}
