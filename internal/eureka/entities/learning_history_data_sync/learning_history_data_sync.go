package entities

import (
	"github.com/jackc/pgtype"
)

type MappingCourseID struct {
	ManabieCourseID pgtype.Text
	WithusCourseID  pgtype.Text
	LastUpdatedDate pgtype.Timestamptz
	LastUpdatedBy   pgtype.Text
	IsArchived      pgtype.Bool
}

type MappingExamLoID struct {
	ExamLoID        pgtype.Text
	MaterialCode    pgtype.Text
	LastUpdatedDate pgtype.Timestamptz
	LastUpdatedBy   pgtype.Text
	IsArchived      pgtype.Bool
}

type MappingQuestionTag struct {
	ManabieTagID    pgtype.Text
	ManabieTagName  pgtype.Text
	WithusTagName   pgtype.Text
	LastUpdatedDate pgtype.Timestamptz
	LastUpdatedBy   pgtype.Text
	IsArchived      pgtype.Bool
}

type FailedSyncEmailRecipient struct {
	RecipientID     pgtype.Text
	EmailAddress    pgtype.Text
	LastUpdatedDate pgtype.Timestamptz
	LastUpdatedBy   pgtype.Text
	IsArchived      pgtype.Bool
}

func (e *MappingCourseID) TableName() string {
	return "withus_mapping_course_id"
}

func (e *MappingExamLoID) TableName() string {
	return "withus_mapping_exam_lo_id"
}

func (e *MappingQuestionTag) TableName() string {
	return "withus_mapping_question_tag"
}

func (e *FailedSyncEmailRecipient) TableName() string {
	return "withus_failed_sync_email_recipient"
}

func (e *MappingCourseID) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"manabie_course_id", "withus_course_id", "last_updated_date", "last_updated_by", "is_archived"}
	values = []interface{}{&e.ManabieCourseID, &e.WithusCourseID, &e.LastUpdatedDate, &e.LastUpdatedBy, &e.IsArchived}
	return
}

func (e *MappingExamLoID) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"exam_lo_id", "material_code", "last_updated_date", "last_updated_by", "is_archived"}
	values = []interface{}{&e.ExamLoID, &e.MaterialCode, &e.LastUpdatedDate, &e.LastUpdatedBy, &e.IsArchived}
	return
}

func (e *MappingQuestionTag) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"manabie_tag_id", "manabie_tag_name", "withus_tag_name", "last_updated_date", "last_updated_by", "is_archived"}
	values = []interface{}{&e.ManabieTagID, &e.ManabieTagName, &e.WithusTagName, &e.LastUpdatedDate, &e.LastUpdatedBy, &e.IsArchived}
	return
}

func (e *FailedSyncEmailRecipient) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"recipient_id", "email_address", "last_updated_date", "last_updated_by", "is_archived"}
	values = []interface{}{&e.RecipientID, &e.EmailAddress, &e.LastUpdatedDate, &e.LastUpdatedBy, &e.IsArchived}
	return
}
