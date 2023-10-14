package helper

import (
	"context"

	"github.com/manabie-com/backend/features/eibanam/communication/util"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

func (h *CommunicationHelper) GetLearnerAppChat(ctx context.Context, token string) (*legacytpb.ConversationListResponse, error) {
	req := legacytpb.ConversationListRequest{
		Limit: 10,
	}

	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	res, err := legacytpb.NewChatServiceClient(h.tomGRPCConn).ConversationList(ctx, &req)
	return res, err
}

func (h *CommunicationHelper) LeaveSupportChatGroup(ctx context.Context, token string, chatID string) error {
	req := &tpb.LeaveConversationsRequest{
		ConversationIds: []string{chatID},
	}
	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	_, err := tpb.NewChatModifierServiceClient(h.tomGRPCConn).LeaveConversations(ctx, req)
	return err
}

func (h *CommunicationHelper) JoinSupportChatGroup(ctx context.Context, token string, chatID string) error {
	req := &tpb.JoinConversationsRequest{
		ConversationIds: []string{chatID},
	}
	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	_, err := tpb.NewChatModifierServiceClient(h.tomGRPCConn).JoinConversations(ctx, req)
	return err
}

func (h *CommunicationHelper) ListSupportJoinedChat(ctx context.Context, token string) (*tpb.ListConversationsInSchoolResponse, error) {
	req := tpb.ListConversationsInSchoolRequest{
		Paging: &cpb.Paging{
			Limit: 20,
		},
		JoinStatus: tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED,
	}

	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	res, err := tpb.NewChatReaderServiceClient(h.tomGRPCConn).ListConversationsInSchool(ctx, &req)
	return res, err
}

func (h *CommunicationHelper) ListSupportUnjoinedChat(ctx context.Context, token string) (*tpb.ListConversationsInSchoolResponse, error) {
	req := tpb.ListConversationsInSchoolRequest{
		Paging: &cpb.Paging{
			Limit: 20,
		},
		JoinStatus: tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_NOT_JOINED,
	}

	ctx, cancel := util.ContextWithTokenAndTimeOut(ctx, token)
	defer cancel()

	res, err := tpb.NewChatReaderServiceClient(h.tomGRPCConn).ListConversationsInSchool(ctx, &req)
	return res, err
}
