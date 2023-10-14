package domain

import "time"

type LessonSearch struct {
	LessonID       string
	LocationID     string
	TeachingMedium string
	TeachingMethod string
	ClassID        string
	CourseID       string

	LessonMember  []*LessonMemberEs
	LessonTeacher []string

	StartTime time.Time
	EndTime   time.Time

	DeletedAt *time.Time
	UpdatedAt time.Time
	CreatedAt time.Time
}

func (l *LessonSearch) AddLessonMembers(lm []*LessonMemberEs) {
	l.LessonMember = lm
}

type LessonSearchs []*LessonSearch

type LessonMemberEs struct {
	ID           string
	Name         string
	CurrentGrade int
	CourseID     string
}
