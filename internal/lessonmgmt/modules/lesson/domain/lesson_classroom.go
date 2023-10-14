package domain

import (
	"fmt"
	"time"
)

type LessonClassroom struct {
	ClassroomID   string
	ClassroomName string
	ClassroomArea string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

func (l LessonClassroom) IsValid() error {
	if len(l.ClassroomID) == 0 {
		return fmt.Errorf("Lesson.Classroom.ClassroomID cannot be empty")
	}
	return nil
}

func (l *LessonClassroom) WithClassroomName(classroomName string) *LessonClassroom {
	l.ClassroomName = classroomName
	return l
}

func (l *LessonClassroom) WithClassroomArea(classroomArea string) *LessonClassroom {
	l.ClassroomArea = classroomArea
	return l
}

type LessonClassrooms []*LessonClassroom

func (l LessonClassrooms) GetIDs() []string {
	ids := make([]string, 0, len(l))
	for _, classroom := range l {
		ids = append(ids, classroom.ClassroomID)
	}
	return ids
}
