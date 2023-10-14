package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Class struct {
	ClassID    pgtype.Text
	Name       pgtype.Text
	CourseID   pgtype.Text
	LocationID pgtype.Text
	SchoolID   pgtype.Text
	UpdatedAt  pgtype.Timestamptz
	CreatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
}

func (c *Class) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at", "deleted_at"}
	values = []interface{}{&c.ClassID, &c.Name, &c.CourseID, &c.LocationID, &c.SchoolID, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt}
	return
}

func (c *Class) TableName() string {
	return "class"
}

func (c *Class) ToClassEntity() *domain.Class {
	class := &domain.Class{
		ClassID:    c.ClassID.String,
		Name:       c.Name.String,
		CourseID:   c.CourseID.String,
		LocationID: c.LocationID.String,
		SchoolID:   c.SchoolID.String,
		CreatedAt:  c.CreatedAt.Time,
		UpdatedAt:  c.UpdatedAt.Time,
	}
	if c.DeletedAt.Status == pgtype.Present {
		class.DeletedAt = &c.DeletedAt.Time
	}
	return class
}

func NewClassFromEntity(c *domain.Class) (*Class, error) {
	classDTO := &Class{}
	database.AllNullEntity(classDTO)
	if err := multierr.Combine(
		classDTO.ClassID.Set(c.ClassID),
		classDTO.Name.Set(c.Name),
		classDTO.CourseID.Set(c.CourseID),
		classDTO.LocationID.Set(c.LocationID),
		classDTO.SchoolID.Set(c.SchoolID),
		classDTO.CreatedAt.Set(c.CreatedAt),
		classDTO.UpdatedAt.Set(c.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from class entity to class dto: %w", err)
	}
	if c.DeletedAt != nil {
		if err := classDTO.DeletedAt.Set(c.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return classDTO, nil
}

type Classes []*Class

func (cc *Classes) Add() database.Entity {
	e := &Class{}
	*cc = append(*cc, e)

	return e
}

type ClassWithCourseStudent struct {
	CourseID  pgtype.Text
	StudentID pgtype.Text
	ClassID   pgtype.Text
}

func (c *ClassWithCourseStudent) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"class_id", "course_id", "student_id",
		}, []interface{}{
			&c.ClassID, &c.CourseID, &c.StudentID,
		}
}

func NewCourseStudentFromEntity(courseStudent *domain.ClassWithCourseStudent) (*ClassWithCourseStudent, error) {
	courseStudentDTO := &ClassWithCourseStudent{}
	if err := multierr.Combine(
		courseStudentDTO.CourseID.Set(courseStudent.CourseID),
		courseStudentDTO.StudentID.Set(courseStudent.StudentID),
		courseStudentDTO.ClassID.Set(courseStudent.ClassID),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from course_student entity to course_student dto")
	}
	return courseStudentDTO, nil
}

func NewCourseStudentFromDto(courseStudent *ClassWithCourseStudent) *domain.ClassWithCourseStudent {
	courseStudentDTO := &domain.ClassWithCourseStudent{}
	courseStudentDTO.CourseID = courseStudent.CourseID.String
	courseStudentDTO.StudentID = courseStudent.StudentID.String
	courseStudentDTO.ClassID = courseStudent.ClassID.String

	return courseStudentDTO
}
