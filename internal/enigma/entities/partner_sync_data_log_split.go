package entities

import (
	"github.com/jackc/pgtype"
)

type PartnerSyncDataLogSplit struct {
	PartnerSyncDataLogSplitID pgtype.Text
	PartnerSyncDataLogID      pgtype.Text
	Payload                   pgtype.JSONB
	Kind                      pgtype.Text
	Status                    pgtype.Text
	RetryTimes                pgtype.Int4
	UpdatedAt                 pgtype.Timestamptz
	CreatedAt                 pgtype.Timestamptz
}

func (p *PartnerSyncDataLogSplit) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"partner_sync_data_log_split_id", "partner_sync_data_log_id", "payload", "kind", "status", "retry_times", "updated_at", "created_at"}
	values = []interface{}{&p.PartnerSyncDataLogSplitID, &p.PartnerSyncDataLogID, &p.Payload, &p.Kind, &p.Status, &p.RetryTimes, &p.UpdatedAt, &p.CreatedAt}
	return
}

func (*PartnerSyncDataLogSplit) TableName() string {
	return "partner_sync_data_log_split"
}

type (
	Kind   string
	Status string
)

const (
	KindStudent        Kind = "STUDENT"
	KindStaff          Kind = "STAFF"
	KindCourse         Kind = "COURSE"
	KindClass          Kind = "CLASS"
	KindLesson         Kind = "LESSON"
	KindAcademicYear   Kind = "ACADEMICYEAR"
	KindStudentLessons Kind = "STUDENT_LESSONS"

	StatusPending    Status = "PENDING"
	StatusProcessing Status = "PROCESSING"
	StatusSuccess    Status = "SUCCESS"
	StatusFailed     Status = "FAILED"
)

var (
	PartnerSyncDataLogStatusValue = map[string]int32{
		"PENDING":    0,
		"PROCESSING": 1,
		"SUCCESS":    2,
		"FAILED":     3,
	}
)

type PartnerSyncDataLogReport struct {
	Status    pgtype.Text
	Total     pgtype.Int8
	CreatedAt pgtype.Date
}
