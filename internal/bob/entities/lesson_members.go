package entities

import (
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
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
	LessonMemberAttendanceNotice = "attendance_notice"
	LessonMemberAttendanceReason = "attendance_reason"
	LessonMemberAttendanceNote   = "attendance_note"
)

type LessonMember struct {
	LessonID         pgtype.Text
	UserID           pgtype.Text
	AttendanceStatus pgtype.Text
	AttendanceRemark pgtype.Text
	CourseID         pgtype.Text
	AttendanceNotice pgtype.Text
	AttendanceReason pgtype.Text
	AttendanceNote   pgtype.Text
	UpdatedAt        pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	DeleteAt         pgtype.Timestamptz
}

func (rcv *LessonMember) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"lesson_id", "user_id", "updated_at", "created_at", "deleted_at", "attendance_status", "attendance_remark", "course_id", "attendance_notice", "attendance_reason", "attendance_note"}
	values = []interface{}{&rcv.LessonID, &rcv.UserID, &rcv.UpdatedAt, &rcv.CreatedAt, &rcv.DeleteAt, &rcv.AttendanceStatus, &rcv.AttendanceRemark, &rcv.CourseID, &rcv.AttendanceNotice, &rcv.AttendanceReason, &rcv.AttendanceNote}
	return
}

func (*LessonMember) TableName() string {
	return "lesson_members"
}

type LessonMembers []*LessonMember

func (u *LessonMembers) Add() database.Entity {
	e := &LessonMember{}
	*u = append(*u, e)

	return e
}

type LessonMemberStates []*LessonMemberState

func (ls *LessonMemberStates) Add() database.Entity {
	e := &LessonMemberState{}
	*ls = append(*ls, e)

	return e
}

func (ls LessonMemberStates) GroupByUserID() map[string]LessonMemberStates {
	res := make(map[string]LessonMemberStates)
	for _, state := range ls {
		userID := state.UserID.String
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
	LessonID         pgtype.Text
	UserID           pgtype.Text
	StateType        pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeleteAt         pgtype.Timestamptz
	BoolValue        pgtype.Bool
	StringArrayValue pgtype.TextArray
}

func (lms *LessonMemberState) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"lesson_id", "user_id", "state_type", "created_at", "updated_at", "deleted_at", "bool_value", "string_array_value"}
	values = []interface{}{&lms.LessonID, &lms.UserID, &lms.StateType, &lms.CreatedAt, &lms.UpdatedAt, &lms.DeleteAt, &lms.BoolValue, &lms.StringArrayValue}
	return
}

func (*LessonMemberState) TableName() string {
	return "lesson_members_states"
}

type StateValue struct {
	BoolValue        pgtype.Bool
	StringArrayValue pgtype.TextArray
}

func (sv *StateValue) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool_value", "string_array_value"}
	values = []interface{}{&sv.BoolValue, &sv.StringArrayValue}
	return
}
