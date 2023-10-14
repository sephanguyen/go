package repo

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Course struct {
	ID             pgtype.Text
	Name           pgtype.Text
	DisplayOrder   pgtype.Int2
	SchoolID       pgtype.Int4
	Icon           pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
	TeachingMethod pgtype.Text
	CourseTypeID   pgtype.Text
	EndDate        pgtype.Timestamptz
	IsArchived     pgtype.Bool
	Remarks        pgtype.Text
	PartnerID      pgtype.Text
}

func (c *Course) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_id",
		"name",
		"display_order",
		"school_id",
		"updated_at",
		"created_at",
		"deleted_at",
		"icon",
		"teaching_method",
		"course_type_id",
		"end_date",
		"is_archived",
		"remarks",
		"course_partner_id",
	}
	values = []interface{}{
		&c.ID,
		&c.Name,
		&c.DisplayOrder,
		&c.SchoolID,
		&c.UpdatedAt,
		&c.CreatedAt,
		&c.DeletedAt,
		&c.Icon,
		&c.TeachingMethod,
		&c.CourseTypeID,
		&c.EndDate,
		&c.IsArchived,
		&c.Remarks,
		&c.PartnerID,
	}
	return
}

func (*Course) TableName() string {
	return "courses"
}

func (c *Course) ToCourseEntity() *domain.Course {
	course := &domain.Course{
		CourseID:       c.ID.String,
		Name:           c.Name.String,
		DisplayOrder:   int(c.DisplayOrder.Int),
		SchoolID:       int(c.SchoolID.Int),
		Icon:           c.Icon.String,
		CreatedAt:      c.CreatedAt.Time,
		UpdatedAt:      c.UpdatedAt.Time,
		EndDate:        c.EndDate.Time,
		TeachingMethod: domain.CourseTeachingMethod(c.TeachingMethod.String),
		CourseTypeID:   c.CourseTypeID.String,
		IsArchived:     c.IsArchived.Bool,
		Remarks:        c.Remarks.String,
		PartnerID:      c.PartnerID.String,
	}
	if c.DeletedAt.Status == pgtype.Present {
		course.DeletedAt = &c.DeletedAt.Time
	}
	return course
}

func NewCourseFromEntity(c *domain.Course) (*Course, error) {
	// HACK: end_date for courses to adapt with join lesson logic
	var courseEndData = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	dto := &Course{}
	database.AllNullEntity(dto)
	if err := multierr.Combine(
		dto.ID.Set(c.CourseID),
		dto.Name.Set(c.Name),
		dto.DisplayOrder.Set(c.DisplayOrder),
		dto.TeachingMethod.Set(c.TeachingMethod),
		dto.SchoolID.Set(c.SchoolID),
		dto.Icon.Set(c.Icon),
		dto.CreatedAt.Set(c.CreatedAt),
		dto.UpdatedAt.Set(c.UpdatedAt),
		dto.DeletedAt.Set(nil),
		dto.EndDate.Set(courseEndData),
		dto.IsArchived.Set(c.IsArchived),
		dto.Remarks.Set(c.Remarks),
		dto.PartnerID.Set(c.PartnerID),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from course entity to course dto: %w", err)
	}
	if len(c.CourseTypeID) > 0 {
		if err := dto.CourseTypeID.Set(c.CourseTypeID); err != nil {
			return nil, fmt.Errorf("could not course_type: %w", err)
		}
	}
	return dto, nil
}
