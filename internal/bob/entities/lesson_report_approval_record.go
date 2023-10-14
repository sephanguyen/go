package entities

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// Lesson Report Approval Record reflect lesson_report_approval_records table
type LessonReportApprovalRecord struct {
	RecordID       pgtype.Text
	LessonReportID pgtype.Text
	Description    pgtype.Text
	ApprovedBy     pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

// FieldMap Lesson Report Approval Record table data fields
func (lr *LessonReportApprovalRecord) FieldMap() ([]string, []interface{}) {
	return []string{
			"record_id",
			"lesson_report_id",
			"description",
			"approved_by",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&lr.RecordID,
			&lr.LessonReportID,
			&lr.Description,
			&lr.ApprovedBy,
			&lr.CreatedAt,
			&lr.UpdatedAt,
			&lr.DeletedAt,
		}
}

// TableName returns "lesson_report_approval_records"
func (lr *LessonReportApprovalRecord) TableName() string {
	return "lesson_report_approval_records"
}

func (lr *LessonReportApprovalRecord) PreUpdate() error {
	return lr.UpdatedAt.Set(time.Now())
}

type LessonReportApprovalRecords []*LessonReportApprovalRecord

func (lr *LessonReportApprovalRecords) Add() database.Entity {
	e := &LessonReportApprovalRecord{}
	*lr = append(*lr, e)

	return e
}
