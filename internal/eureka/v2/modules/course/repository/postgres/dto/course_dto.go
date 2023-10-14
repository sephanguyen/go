package dto

import (
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type CourseDto struct {
	ID             pgtype.Text
	Name           pgtype.Text
	Icon           pgtype.Text
	DisplayOrder   pgtype.Int2
	SchoolID       pgtype.Int4
	PartnerID      pgtype.Text
	Remarks        pgtype.Text
	TeachingMethod pgtype.Text
	CourseTypeID   pgtype.Text
	IsArchived     pgtype.Bool

	BookID    pgtype.Text
	StartDate pgtype.Timestamptz
	EndDate   pgtype.Timestamptz

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c CourseDto) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"course_id", "name", "icon", "display_order", "school_id", "course_partner_id", "remarks", "teaching_method", "course_type_id", "is_archived", "start_date", "end_date", "created_at", "updated_at", "deleted_at", "book_id"}
	values = []interface{}{&c.ID, &c.Name, &c.Icon, &c.DisplayOrder, &c.SchoolID, &c.PartnerID, &c.Remarks, &c.TeachingMethod, &c.CourseTypeID, &c.IsArchived, &c.StartDate, &c.EndDate, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.BookID}
	return
}

func (c CourseDto) TableName() string {
	return "courses"
}

type CourseDtos []*CourseDto

func (u *CourseDtos) Add() database.Entity {
	e := &CourseDto{}
	*u = append(*u, e)
	return e
}

func (c CourseDto) ToCourseEntity() *domain.Course {
	course := domain.Course{
		ID:           c.ID.String,
		Name:         c.Name.String,
		Icon:         c.Icon.String,
		DisplayOrder: int(c.DisplayOrder.Int),
		SchoolID:     int(c.SchoolID.Int),
		PartnerID:    c.PartnerID.String,

		Remarks:        c.Remarks.String,
		TeachingMethod: c.TeachingMethod.String,
		CourseTypeID:   c.CourseTypeID.String,
		IsArchived:     c.IsArchived.Bool,

		StartDate: c.StartDate.Time,
		EndDate:   c.EndDate.Time,

		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,

		BookID: c.BookID.String,
	}
	if c.DeletedAt.Status == pgtype.Present {
		course.DeletedAt = &c.DeletedAt.Time
	}
	return &course
}
