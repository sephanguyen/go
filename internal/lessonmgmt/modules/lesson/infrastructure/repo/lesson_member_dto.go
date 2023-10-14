package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
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
	DeletedAt        pgtype.Timestamptz
	// First name == Given Name on UI
	UserFirstName pgtype.Text
	UserLastName  pgtype.Text
}

func (l *LessonMember) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"lesson_id",
		"user_id",
		"updated_at",
		"created_at",
		"deleted_at",
		"attendance_status",
		"attendance_remark",
		"course_id",
		"attendance_notice",
		"attendance_reason",
		"attendance_note",
		"user_first_name",
		"user_last_name",
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
		&l.AttendanceNotice,
		&l.AttendanceReason,
		&l.AttendanceNote,
		&l.UserFirstName,
		&l.UserLastName,
	}
	return
}

func (*LessonMember) TableName() string {
	return "lesson_members"
}

func (l *LessonMember) PreUpsert() error {
	now := time.Now()
	if l.CreatedAt.Status != pgtype.Present {
		if err := l.CreatedAt.Set(now); err != nil {
			return err
		}
	}
	if err := l.UpdatedAt.Set(now); err != nil {
		return err
	}

	return nil
}

func NewLessonMembersFromLessonEntity(l *domain.Lesson) (LessonMembers, error) {
	dtos := make(LessonMembers, 0, len(l.Learners))
	for _, learner := range l.Learners {
		dto := &LessonMember{}
		database.AllNullEntity(dto)
		if err := multierr.Combine(
			dto.LessonID.Set(l.LessonID),
			dto.CourseID.Set(learner.CourseID),
			dto.UserID.Set(learner.LearnerID),
			dto.AttendanceStatus.Set(learner.AttendStatus),
			dto.AttendanceNotice.Set(learner.AttendanceNotice),
			dto.AttendanceReason.Set(learner.AttendanceReason),
		); err != nil {
			return nil, fmt.Errorf("could not mapping from lesson members of lesson entity to lesson members dto: %w", err)
		}

		if len(learner.AttendanceNote) > 0 {
			if err := dto.AttendanceNote.Set(learner.AttendanceNote); err != nil {
				return nil, fmt.Errorf("failed to set attendance note of lesson entity to lesson members dto: %w", err)
			}
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}

func NewLessonMembersFromLessonMemberEntity(lm *domain.LessonMember) (*LessonMember, error) {
	dto := &LessonMember{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.LessonID.Set(lm.LessonID),
		dto.CourseID.Set(lm.CourseID),
		dto.UserID.Set(lm.StudentID),
		dto.AttendanceStatus.Set(lm.AttendanceStatus),
		dto.AttendanceRemark.Set(lm.AttendanceRemark),
		dto.UserLastName.Set(lm.UserFirstName),
		dto.UserFirstName.Set(lm.UserLastName),
		dto.CreatedAt.Set(lm.CreatedAt),
		dto.UpdatedAt.Set(lm.UpdatedAt),
		dto.DeletedAt.Set(lm.DeletedAt),
		dto.AttendanceNote.Set(lm.AttendanceNote),
		dto.AttendanceNotice.Set(lm.AttendanceNotice),
		dto.AttendanceReason.Set(lm.AttendanceReason),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from lesson members of lesson entity to lesson members dto: %w", err)
	}
	return dto, nil
}

type LessonMembers []*LessonMember

func (u *LessonMembers) Add() database.Entity {
	e := &LessonMember{}
	*u = append(*u, e)

	return e
}

func (l *LessonMember) ToLessonMemberEntity() *domain.LessonMember {
	lm := &domain.LessonMember{
		LessonID:         l.LessonID.String,
		StudentID:        l.UserID.String,
		AttendanceStatus: l.AttendanceStatus.String,
		AttendanceRemark: l.AttendanceRemark.String,
		CourseID:         l.CourseID.String,
		AttendanceNotice: l.AttendanceNotice.String,
		AttendanceReason: l.AttendanceReason.String,
		AttendanceNote:   l.AttendanceNote.String,
		UserFirstName:    l.UserFirstName.String,
		UserLastName:     l.UserFirstName.String,
		CreatedAt:        l.CreatedAt.Time,
		UpdatedAt:        l.UpdatedAt.Time,
	}
	if l.DeletedAt.Status == pgtype.Present {
		lm.DeletedAt = &l.DeletedAt.Time
	}
	return lm
}

func (l *LessonMember) ToLessonLearnerEntity() *domain.LessonLearner {
	lessonLearner := &domain.LessonLearner{
		LearnerID:      l.UserID.String,
		CourseID:       l.CourseID.String,
		AttendanceNote: l.AttendanceNote.String,
	}

	if l.AttendanceStatus.Status == pgtype.Present {
		lessonLearner.AttendStatus = domain.StudentAttendStatus(l.AttendanceStatus.String)
	}

	if l.AttendanceNotice.Status == pgtype.Present {
		lessonLearner.AttendanceNotice = domain.StudentAttendanceNotice(l.AttendanceNotice.String)
	}

	if l.AttendanceReason.Status == pgtype.Present {
		lessonLearner.AttendanceReason = domain.StudentAttendanceReason(l.AttendanceReason.String)
	}

	return lessonLearner
}
