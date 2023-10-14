package domain

import (
	"time"
)

type (
	UpdateLessonMemberField  string
	UpdateLessonMemberFields []UpdateLessonMemberField
)

func (u UpdateLessonMemberFields) StringArray() []string {
	res := make([]string, 0, len(u))
	for _, f := range u {
		res = append(res, string(f))
	}

	return res
}

const (
	LessonMemberCourseID         = "course_id"
	LessonMemberAttendanceStatus = "attendance_status"
	LessonMemberAttendanceRemark = "attendance_remark"
)

type LessonMember struct {
	LessonID         string
	UserID           string
	AttendanceStatus string
	AttendanceRemark string
	CourseID         string
	UpdatedAt        time.Time
	CreatedAt        time.Time
	DeletedAt        *time.Time
}

type LessonMembers []*LessonMember

type LessonMemberStates []*LessonMemberState

func (ls LessonMemberStates) GroupByUserID() map[string]LessonMemberStates {
	res := make(map[string]LessonMemberStates)
	for _, state := range ls {
		userID := state.UserID
		if v, ok := res[userID]; !ok {
			res[userID] = LessonMemberStates{state}
		} else {
			v = append(v, state)
			res[userID] = v
		}
	}

	return res
}

type LessonMemberState struct {
	LessonID         string
	UserID           string
	AttendanceStatus string
	AttendanceRemark string
	CourseID         string
	UpdatedAt        time.Time
	CreatedAt        time.Time
	DeletedAt        *time.Time
	StateType        string
	BoolValue        bool
	StringArrayValue []string
}
