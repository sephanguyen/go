package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type ConversionTask struct {
	TaskUUID           pgtype.Text
	ResourceURL        pgtype.Text
	Status             pgtype.Text
	ConversionResponse pgtype.JSONB
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
}

func (c *ConversionTask) FieldMap() ([]string, []interface{}) {
	return []string{
			"task_uuid", "resource_url", "status", "conversion_response", "created_at", "updated_at",
		}, []interface{}{
			&c.TaskUUID, &c.ResourceURL, &c.Status, &c.ConversionResponse, &c.CreatedAt, &c.UpdatedAt,
		}
}

func (c *ConversionTask) TableName() string {
	return "conversion_tasks"
}

type ConversionTasks []*ConversionTask

func (e *ConversionTasks) Add() database.Entity {
	t := &ConversionTask{}
	*e = append(*e, t)

	return t
}
