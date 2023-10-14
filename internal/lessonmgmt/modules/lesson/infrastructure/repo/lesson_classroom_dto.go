package repo

import (
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
)

type LessonClassroom struct {
	LessonID    pgtype.Text
	ClassroomID pgtype.Text
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (l *LessonClassroom) FieldMap() ([]string, []interface{}) {
	return []string{
			"lesson_id",
			"classroom_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&l.LessonID,
			&l.ClassroomID,
			&l.CreatedAt,
			&l.UpdatedAt,
			&l.DeletedAt,
		}
}

func (l *LessonClassroom) TableName() string {
	return "lesson_classrooms"
}

func (l *LessonClassroom) ToLessonClassroomEntity() *domain.LessonClassroom {
	lessonClassroom := &domain.LessonClassroom{
		ClassroomID: l.ClassroomID.String,
	}

	return lessonClassroom
}
