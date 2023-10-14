package domain

import "time"

type LessonMember struct {
	LessonID         string
	StudentID        string
	AttendanceStatus string
	AttendanceRemark string
	CourseID         string
	AttendanceReason string
	AttendanceNotice string
	AttendanceNote   string
	UserFirstName    string
	UserLastName     string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func (lm *LessonMember) GetStudentCourse() string {
	return lm.StudentID + "-" + lm.CourseID
}

func (lm *LessonMember) GetKey() string {
	return lm.LessonID + "-" + lm.StudentID
}

type LessonMembers []*LessonMember

func (l LessonMembers) GetStudentIDs() []string {
	ids := make([]string, 0, len(l))
	for _, member := range l {
		ids = append(ids, member.StudentID)
	}
	return ids
}

func (l LessonMembers) GetMapFieldValuesOfStudent() map[string]*LessonMember {
	mapFieldValuesOfStudent := make(map[string]*LessonMember)
	for _, lessonMember := range l {
		mapFieldValuesOfStudent[lessonMember.StudentID] = lessonMember
	}
	return mapFieldValuesOfStudent
}
