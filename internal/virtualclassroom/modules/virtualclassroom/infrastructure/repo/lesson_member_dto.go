package repo

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
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

type LessonMemberDTO struct {
	LessonID         pgtype.Text
	UserID           pgtype.Text
	AttendanceStatus pgtype.Text
	AttendanceRemark pgtype.Text
	CourseID         pgtype.Text
	UpdatedAt        pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func (l *LessonMemberDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"lesson_id",
		"user_id",
		"updated_at",
		"created_at",
		"deleted_at",
		"attendance_status",
		"attendance_remark",
		"course_id",
	}
	values = []interface{}{
		&l.LessonID,
		&l.UserID,
		&l.UpdatedAt,
		&l.CreatedAt,
		&l.DeletedAt,
		&l.AttendanceStatus,
		&l.AttendanceRemark,
		&l.CourseID,
	}
	return
}

func (l *LessonMemberDTO) TableName() string {
	return "lesson_members"
}

func (l *LessonMemberDTO) ToLessonMemberDomain() domain.LessonMember {
	domain := domain.LessonMember{
		LessonID:         l.LessonID.String,
		UserID:           l.UserID.String,
		AttendanceStatus: l.AttendanceStatus.String,
		AttendanceRemark: l.AttendanceRemark.String,
		CourseID:         l.CourseID.String,
		CreatedAt:        l.CreatedAt.Time,
		UpdatedAt:        l.UpdatedAt.Time,
	}

	if l.DeletedAt.Status == pgtype.Present {
		domain.DeletedAt = &l.DeletedAt.Time
	}

	return domain
}

type LessonMemberDTOs []*LessonMemberDTO

func (u *LessonMemberDTOs) Add() database.Entity {
	e := &LessonMemberDTO{}
	*u = append(*u, e)

	return e
}

type LessonMemberStateDTOs []*LessonMemberStateDTO

func (ls *LessonMemberStateDTOs) Add() database.Entity {
	e := &LessonMemberStateDTO{}
	*ls = append(*ls, e)

	return e
}

func (ls *LessonMemberStateDTOs) ToLessonMemberStatesDomainEntity() domain.LessonMemberStates {
	lms := make(domain.LessonMemberStates, 0, len(*ls))

	for _, lmsDTO := range *ls {
		strArrayValue := make([]string, 0, len(lmsDTO.StringArrayValue.Elements))
		for _, str := range lmsDTO.StringArrayValue.Elements {
			strArrayValue = append(strArrayValue, str.String)
		}

		lmsDomain := &domain.LessonMemberState{
			LessonID:         lmsDTO.LessonID.String,
			UserID:           lmsDTO.UserID.String,
			StateType:        lmsDTO.StateType.String,
			BoolValue:        lmsDTO.BoolValue.Bool,
			StringArrayValue: strArrayValue,
			CreatedAt:        lmsDTO.CreatedAt.Time,
			UpdatedAt:        lmsDTO.UpdatedAt.Time,
		}
		if lmsDTO.DeletedAt.Status == pgtype.Present {
			lmsDomain.DeletedAt = &lmsDTO.DeletedAt.Time
		}

		lms = append(lms, lmsDomain)
	}

	return lms
}

func (ls LessonMemberStateDTOs) GroupByUserID() map[string]LessonMemberStateDTOs {
	res := make(map[string]LessonMemberStateDTOs)
	for _, state := range ls {
		userID := state.UserID.String
		if v, ok := res[userID]; !ok {
			res[userID] = LessonMemberStateDTOs{state}
		} else {
			v = append(v, state)
			res[userID] = v
		}
	}

	return res
}

type LessonMemberStateDTO struct {
	LessonID         pgtype.Text
	UserID           pgtype.Text
	StateType        pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	BoolValue        pgtype.Bool
	StringArrayValue pgtype.TextArray
}

func (lms *LessonMemberStateDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"lesson_id",
		"user_id",
		"state_type",
		"created_at",
		"updated_at",
		"deleted_at",
		"bool_value",
		"string_array_value",
	}
	values = []interface{}{
		&lms.LessonID,
		&lms.UserID,
		&lms.StateType,
		&lms.CreatedAt,
		&lms.UpdatedAt,
		&lms.DeletedAt,
		&lms.BoolValue,
		&lms.StringArrayValue,
	}
	return
}

func (*LessonMemberStateDTO) TableName() string {
	return "lesson_members_states"
}

type StateValueDTO struct {
	BoolValue        pgtype.Bool
	StringArrayValue pgtype.TextArray
}

func (sv *StateValueDTO) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"bool_value", "string_array_value"}
	values = []interface{}{&sv.BoolValue, &sv.StringArrayValue}
	return
}
