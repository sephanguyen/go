package repo

import (
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
)

type CourseTeachingTimeToExport struct {
	CourseID        pgtype.Text
	Name            pgtype.Text
	PreparationTime pgtype.Int4
	BreakTime       pgtype.Int4
	UpdatedAt       pgtype.Timestamptz
	CreatedAt       pgtype.Timestamptz
	DeletedAt       pgtype.Timestamptz
}

func (*CourseTeachingTimeToExport) TableName() string {
	return "course_teaching_time"
}

func (c *CourseTeachingTimeToExport) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "name", "preparation_time", "break_time", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.CourseID, &c.Name, &c.PreparationTime, &c.BreakTime, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (c *CourseTeachingTimeToExport) ToCourseEntity() *domain.Course {
	return &domain.Course{
		CourseID:        c.CourseID.String,
		Name:            c.Name.String,
		PreparationTime: c.PreparationTime.Int,
		BreakTime:       c.BreakTime.Int,
		CreatedAt:       c.CreatedAt.Time,
		UpdatedAt:       c.UpdatedAt.Time,
	}
}
