package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Class struct {
	ClassID  pgtype.Text
	CourseID pgtype.Text
}

func (c *Class) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"class_id", "course_id"}
	values = []interface{}{&c.ClassID, &c.CourseID}
	return
}

func (c *Class) TableName() string {
	return "class"
}

type ClassUnassigned struct {
	StudentSubscriptionID string
	IsClassUnAssigned     bool
}

type ClassRepository interface {
	GetByStudentCourse(ctx context.Context, db database.Ext, studentWithCourse []string) (map[string]string, error)
	GetReserveClass(ctx context.Context, db database.Ext, studentWithCourse []string) (map[string]string, error)
}

type ClassUseCase interface {
	GetByStudentSubscription(ctx context.Context, db database.Ext, studentSubID []string) ([]*ClassUnassigned, error)
}
