package core

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

const (
	ConversationRoleStudent = "USER_GROUP_STUDENT"
	ConversationRoleTeacher = "USER_GROUP_TEACHER"
	ConversationRoleParent  = "USER_GROUP_PARENT"

	ConversationStatusActive   = "CONVERSATION_STATUS_ACTIVE"
	ConversationStatusInActive = "CONVERSATION_STATUS_INACTIVE"
)

type ConversationMembers struct {
	ID             pgtype.Text
	UserID         pgtype.Text
	ConversationID pgtype.Text
	Role           pgtype.Text
	Status         pgtype.Text
	SeenAt         pgtype.Timestamptz
	LastNotifyAt   pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
}

func (e *ConversationMembers) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"conversation_statuses_id", "user_id", "conversation_id", "role", "status", "seen_at", "last_notify_at", "created_at", "updated_at"}
	values = []interface{}{&e.ID, &e.UserID, &e.ConversationID, &e.Role, &e.Status, &e.SeenAt, &e.LastNotifyAt, &e.CreatedAt, &e.UpdatedAt}
	return
}

func (*ConversationMembers) TableName() string {
	return "conversation_members"
}
func CreateConversationMember(userID, conversationID, role string) (*ConversationMembers, error) {
	e := &ConversationMembers{}
	now := time.Now()
	err := multierr.Combine(
		e.ID.Set(idutil.ULIDNow()),
		e.UserID.Set(userID),
		e.ConversationID.Set(conversationID),
		e.Role.Set(role),
		e.Status.Set(ConversationStatusActive),
		e.SeenAt.Set(time.Now()),
		e.LastNotifyAt.Set(nil),
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	return e, err
}

type ListConversationUnjoinedFilter struct {
	UserID      pgtype.Text
	OwnerIDs    pgtype.TextArray
	AccessPaths pgtype.TextArray
}
