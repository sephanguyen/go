package domain

type Reallocation struct {
	OriginalLessonID string
	StudentID        string
	NewLessonID      string
	CourseID         string
}

func (r *Reallocation) GetKey() string {
	return r.OriginalLessonID + "-" + r.StudentID
}
