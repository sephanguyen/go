package support

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	tompb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toMessageResponse(m *entities.Message) *tompb.MessageResponse {
	createdAt, _ := types.TimestampProto(m.CreatedAt.Time)
	isDeleted := !m.DeletedAt.Time.IsZero()

	var content string
	var urlMedia string

	if !isDeleted {
		content = m.Message.String
		urlMedia = m.UrlMedia.String
	}

	return &tompb.MessageResponse{
		MessageId:      m.ID.String,
		ConversationId: m.ConversationID.String,
		UserId:         m.UserID.String,
		Content:        content,
		UrlMedia:       urlMedia,
		Type:           tompb.MessageType(tompb.MessageType_value[m.Type.String]),
		CreatedAt:      createdAt,
		TargetUser:     m.TargetUser.String,
		IsDeleted:      isDeleted,
		DeletedBy:      m.DeletedBy.String,
	}
}

func convertToMapMessages(ms []*entities.Message) map[pgtype.Text]*entities.Message {
	mapMessage := make(map[pgtype.Text]*entities.Message)
	for _, m := range ms {
		mapMessage[m.ConversationID] = m
	}
	return mapMessage
}
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
		Type:           tpb.MessageType(tompb.MessageType_value[m.Type.String]),
		CreatedAt:      timestamppb.New(m.CreatedAt.Time),
		TargetUser:     m.TargetUser.String,
		IsDeleted:      isDeleted,
		DeletedBy:      m.DeletedBy.String,
		UpdatedAt:      timestamppb.New(m.UpdatedAt.Time),
	}
}
func retrieveConversationIDsFromConversation(cs []*entities.Conversation) []string {
	res := make([]string, 0, len(cs))
	for _, c := range cs {
		res = append(res, c.ID.String)
	}
	return res
}

func getStudentIDAndSeenStatusFromConversationMembers(members []*entities.ConversationMembers, message *entities.Message, userID string) (studentID string, seen bool) {
	for _, cStatus := range members {
		if cStatus.UserID.String == userID {
			seen = cStatus.SeenAt.Time.After(message.CreatedAt.Time)
		}
		if cStatus.Role.String == entities.ConversationRoleStudent {
			studentID = cStatus.UserID.String
		}
	}
	return
}
func retrieveConversationIDsFromConversationMembers(cms []*entities.ConversationMembers) []string {
	conversationIDs := make([]string, 0, len(cms))
	for _, cm := range cms {
		conversationIDs = append(conversationIDs, cm.ConversationID.String)
	}
	return conversationIDs
}

func userInfoToEntity(req *upb.EvtUserInfo) (*entities.UserDeviceToken, error) {
	user := &entities.UserDeviceToken{}
	database.AllNullEntity(user)

	user.UserID = database.Text(req.UserId)

	var err error
	if req.DeviceToken != "" {
		err = multierr.Combine(
			user.Token.Set(req.DeviceToken),
			user.AllowNotification.Set(req.AllowNotification),
		)
	}

	if req.Name != "" {
		err = multierr.Append(
			err,
			user.UserName.Set(req.Name),
		)
	}

	return user, err
}
