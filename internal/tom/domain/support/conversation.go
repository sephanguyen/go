package support

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/types"

	"github.com/jackc/pgtype"
)

type ConversationStudent struct {
	ID               pgtype.Text
	ConversationID   pgtype.Text
	StudentID        pgtype.Text
	ConversationType pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	SearchIndexTime  pgtype.Timestamptz
}

func (c *ConversationStudent) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"id", "conversation_id", "student_id", "conversation_type", "created_at", "updated_at", "deleted_at", "search_index_time"}
	values = []interface{}{&c.ID, &c.ConversationID, &c.StudentID, &c.ConversationType, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt, &c.SearchIndexTime}
	return
}

func (*ConversationStudent) TableName() string {
	return "conversation_students"
}

type ConversationFilter struct {
	UserID              string
	RepliedStatus       types.NullBool
	JoinStatus          types.NullBool
	AccessPaths         types.NullStrArr
	ConversationName    types.NullStr
	School              types.NullStrArr
	ConversationTypes   types.NullStrArr
	Courses             types.NullStrArr
	SortBy              []ConversationSortItem
	Limit               types.NullInt64
	OffsetTime          types.NullInt64
	OffsetConverstionID types.NullStr

	// each LocationConfig will hold a list of location ID
	LocationConfigs []LocationConfigFilter
}

type LocationConfigFilter struct {
	ConversationType types.NullStr
	AccessPaths      types.NullStrArr
}

type ConversationSortItem struct {
	Key ConversationSortKey
	Asc bool
}
type ConversationSortKey int

const (
	SortKey_LatestMsgTime ConversationSortKey = iota
	SortKey_ConversationID
)

var (
	DefaultConversationSorts = []ConversationSortItem{
		{
			Key: SortKey_LatestMsgTime,
		},
		{
			Key: SortKey_ConversationID,
			Asc: true,
		},
	}
)

type SearchLastMessage struct {
	Text      string    `json:"text,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SearchConversationDoc struct {
	ConversationID           string            `json:"conversation_id,omitempty"`
	ConversationNameEnglish  string            `json:"conversation_name.english,omitempty"`
	ConversationNameJapanese string            `json:"conversation_name.japanese,omitempty"`
	CourseIDs                []string          `json:"course_ids"`
	UserIDs                  []string          `json:"user_ids"`
	LastMessage              SearchLastMessage `json:"last_message,omitempty"`
	IsReplied                bool              `json:"is_replied"`
	Owner                    string            `json:"owner"`
	ConversationType         string            `json:"conversation_type"`
	AccessPath               []string          `json:"access_paths"`
}

func (d SearchConversationDoc) GetFields() []interface{} {
	return []interface{}{
		d.ConversationID,
		d.ConversationNameEnglish,
		d.ConversationNameJapanese,
		d.CourseIDs,
		d.UserIDs,
		d.LastMessage.Text,
		d.LastMessage.UpdatedAt,
		d.IsReplied,
		d.Owner,
		d.ConversationType,
	}
}

// TODO: use elasticsearch term and seq-no for optimistic concurrency
// https://www.elastic.co/guide/en/elasticsearch/reference/current/optimistic-concurrency-control.html
type UpdateItems struct {
	ConversationID          string
	LatestMessageUpdateTime *time.Time
	Replied                 types.NullBool
	Courses                 types.NullStrArr
	UserIDs                 types.NullStrArr
}
