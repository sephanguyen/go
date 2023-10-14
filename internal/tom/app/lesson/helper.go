package lesson

import (
	"sort"
	"strings"

	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	lessonIgnoredSystemMessages = []string{
		tpb.CodesMessageType_CODES_MESSAGE_TYPE_JOINED_LESSON.String(),
		tpb.CodesMessageType_CODES_MESSAGE_TYPE_LEFT_LESSON.String(),
		tpb.CodesMessageType_CODES_MESSAGE_TYPE_END_LIVE_LESSON.String(),
	}
)

func toMessagePb(m *entities.Message) *tpb.MessageResponse {
	isDeleted := !m.DeletedAt.Time.IsZero()

	var content string
	var urlMedia string

	if !isDeleted {
		content = m.Message.String
		urlMedia = m.UrlMedia.String
	}
	return &tpb.MessageResponse{
		MessageId:      m.ID.String,
		ConversationId: m.ConversationID.String,
		UserId:         m.UserID.String,
		Content:        content,
		UrlMedia:       urlMedia,
		Type:           tpb.MessageType(tpb.MessageType_value[m.Type.String]),
		CreatedAt:      timestamppb.New(m.CreatedAt.Time),
		TargetUser:     m.TargetUser.String,
		IsDeleted:      isDeleted,
		DeletedBy:      m.DeletedBy.String,
		UpdatedAt:      timestamppb.New(m.UpdatedAt.Time),
	}
}

func getFlattenUserIdsByAscending(ids []string) string {
	newIds := make([]string, len(ids))
	copy(newIds, ids)
	sort.Strings(newIds)
	return strings.Join(newIds, "_")
}
