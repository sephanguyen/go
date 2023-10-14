package domain

import "fmt"

type LessonTeacher struct {
	TeacherID string
	Name      string
}

func (l LessonTeacher) Validate() error {
	if len(l.TeacherID) == 0 {
		return fmt.Errorf("Lesson.Teacher.TeacherID cannot be empty")
	}
	return nil
}

type LessonTeachers []*LessonTeacher

func (l LessonTeachers) IsValid() error {
	for i := range l {
		if err := l[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (l LessonTeachers) GetIDs() []string {
	ids := make([]string, 0, len(l))
	for _, u := range l {
		ids = append(ids, u.TeacherID)
	}
	return ids
}
