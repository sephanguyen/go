package domain

import "time"

type Course struct {
	CourseID        string
	Name            string
	PreparationTime int32
	BreakTime       int32
	UpdatedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       time.Time
}

type Courses []*Course

func (c Courses) GetCourseIDs() []string {
	ids := make([]string, 0, len(c))
	for _, u := range c {
		ids = append(ids, u.CourseID)
	}
	return ids
}
