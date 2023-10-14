package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/commands"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VirtualClassroomChatService struct {
	ChatServiceCommand commands.ChatServiceCommand
	Logger             *zap.Logger
}

func (v *VirtualClassroomChatService) GetConversationID(ctx context.Context, req *vpb.GetConversationIDRequest) (*vpb.GetConversationIDResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "lesson ID can't empty")
	}
	participants := req.GetParticipantList()
	if len(participants) == 0 {
		return nil, status.Error(codes.InvalidArgument, "participant list should contain at least one user ID")
	}
	conversationType := domain.LiveLessonConversationType(req.GetConversationType().String())

	conversationID, err := v.ChatServiceCommand.GetConversationID(ctx, req.LessonId, participants, conversationType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.GetConversationIDResponse{
		ConversationId: conversationID,
	}, nil
}

func (v *VirtualClassroomChatService) GetPrivateConversationIDs(ctx context.Context, req *vpb.GetPrivateConversationIDsRequest) (*vpb.GetPrivateConversationIDsResponse, error) {
	if len(strings.TrimSpace(req.LessonId)) == 0 {
		return nil, status.Error(codes.InvalidArgument, "lesson ID can't empty")
	}
	participantIDs := req.GetParticipantIds()
	if len(participantIDs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "participant ids should contain at least one user ID")
	}

	participantConvMap, failedParticipantIDs, err := v.ChatServiceCommand.GetPrivateConversationIDs(ctx, req.LessonId, participantIDs)
	if len(failedParticipantIDs) > 0 {
		v.Logger.Warn("some participant IDs have failed to create private conversation",
			zap.String("error_msg", err.Error()),
			zap.String("lesson_id", req.LessonId),
			zap.String("failed_participants", fmt.Sprintf("%s", failedParticipantIDs)),
		)

		return &vpb.GetPrivateConversationIDsResponse{
			ParticipantConversationMap: participantConvMap,
			FailedPrivConv: &vpb.GetPrivateConversationIDsResponse_FailedPrivateConversation{
				LessonId:       req.LessonId,
				ParticipantIds: failedParticipantIDs,
				ErrorMsg:       err.Error(),
			},
		}, nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &vpb.GetPrivateConversationIDsResponse{
		ParticipantConversationMap: participantConvMap,
	}, nil
}
