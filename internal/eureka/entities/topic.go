package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type (
	TopicType   string
	TopicStatus string
)

const (
	TopicTypeNone       TopicType = "TOPIC_TYPE_NONE"
	TopicTypeLearning   TopicType = "TOPIC_TYPE_LEARNING"
	TopicTypePractical  TopicType = "TOPIC_TYPE_PRACTICAL"
	TopicTypeExam       TopicType = "TOPIC_TYPE_EXAM"
	TopicTypeAssignment TopicType = "TOPIC_TYPE_ASSIGNMENT"
	TopicTypeLiveLesson TopicType = "TOPIC_TYPE_LIVE_LESSON"

	TopicStatusNone      TopicStatus = "TOPIC_STATUS_NONE"
	TopicStatusDraft     TopicStatus = "TOPIC_STATUS_DRAFT"
	TopicStatusPublished TopicStatus = "TOPIC_STATUS_PUBLISHED"
)

// Topic reflect topics table
type Topic struct {
	ID                    pgtype.Text `sql:"topic_id,pk"`
	Name                  pgtype.Text
	Country               pgtype.Text
	Grade                 pgtype.Int2
	Subject               pgtype.Text
	TopicType             pgtype.Text
	Status                pgtype.Text
	ChapterID             pgtype.Text `sql:"chapter_id"`
	DisplayOrder          pgtype.Int2
	IconURL               pgtype.Text `sql:"icon_url"`
	SchoolID              pgtype.Int4 `sql:"school_id"`
	TotalLOs              pgtype.Int4 `sql:"total_los"`
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	PublishedAt           pgtype.Timestamptz
	AttachmentNames       pgtype.TextArray
	AttachmentURLs        pgtype.TextArray `sql:"attachment_urls"`
	Instruction           pgtype.Text
	CopiedTopicID         pgtype.Text `sql:"copied_topic_id"`
	EssayRequired         pgtype.Bool
	DeletedAt             pgtype.Timestamptz
	LODisplayOrderCounter pgtype.Int4
}

// FieldMap topics table data fields
func (t *Topic) FieldMap() ([]string, []interface{}) {
	return []string{
			"topic_id",
			"name",
			"country",
			"grade",
			"subject",
			"topic_type",
			"updated_at",
			"created_at",
			"status",
			"chapter_id",
			"display_order",
			"icon_url",
			"school_id",
			"total_los",
			"published_at",
			"attachment_names",
			"attachment_urls",
			"instruction",
			"copied_topic_id",
			"essay_required",
			"deleted_at",
			"lo_display_order_counter",
		}, []interface{}{
			&t.ID,
			&t.Name,
			&t.Country,
			&t.Grade,
			&t.Subject,
			&t.TopicType,
			&t.UpdatedAt,
			&t.CreatedAt,
			&t.Status,
			&t.ChapterID,
			&t.DisplayOrder,
			&t.IconURL,
			&t.SchoolID,
			&t.TotalLOs,
			&t.PublishedAt,
			&t.AttachmentNames,
			&t.AttachmentURLs,
			&t.Instruction,
			&t.CopiedTopicID,
			&t.EssayRequired,
			&t.DeletedAt,
			&t.LODisplayOrderCounter,
		}
}

// TableName returns "topics"
func (t *Topic) TableName() string {
	return "topics"
}

type Topics []*Topic

func (u *Topics) Add() database.Entity {
	e := &Topic{}
	*u = append(*u, e)

	return e
}

type CopiedTopic struct {
	ID         pgtype.Text
	CopyFromID pgtype.Text
}

type BookTopic struct {
	Topic
	BookID pgtype.Text
}
