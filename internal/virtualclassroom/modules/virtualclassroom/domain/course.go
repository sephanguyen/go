package domain

import "time"

type CourseStatus string

const (
	StatusNone      CourseStatus = "COURSE_STATUS_NONE"
	StatusActive    CourseStatus = "COURSE_STATUS_ACTIVE"
	StatusCompleted CourseStatus = "COURSE_STATUS_COMPLETED"
	StatusOnGoing   CourseStatus = "COURSE_STATUS_ON_GOING"
)

type Course struct {
	ID        string
	Name      string
	Status    string
	UpdatedAt time.Time
	CreatedAt time.Time
	DeletedAt time.Time
}

type Courses []*Course

func (c Courses) GetIDs() []string {
	ids := make([]string, 0, len(c))
	for _, course := range c {
		ids = append(ids, course.ID)
	}
	return ids
}
