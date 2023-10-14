package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/tom/app/core"
	"github.com/manabie-com/backend/internal/tom/app/lesson"
	"github.com/manabie-com/backend/internal/tom/app/support"

	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

type ManabufV1ConversationReader struct {
	SupportChatReader *support.ChatReader
	LessonChatReader  *lesson.ChatReader
	tpb.UnimplementedConversationReaderServiceServer
}

func (rcv *ManabufV1ConversationReader) ListConversationByUsers(ctx context.Context, req *tpb.ListConversationByUsersRequest) (*tpb.ListConversationByUsersResponse, error) {
	return rcv.SupportChatReader.ListConversationByUsers(ctx, req)
}

func (rcv *ManabufV1ConversationReader) ListConversationIDs(ctx context.Context, req *tpb.ListConversationIDsRequest) (*tpb.ListConversationIDsResponse, error) {
	return rcv.SupportChatReader.ListConversationIDs(ctx, req)
}

func (rcv *ManabufV1ConversationReader) ListConversationByLessons(ctx context.Context, req *tpb.ListConversationByLessonsRequest) (*tpb.ListConversationByLessonsResponse, error) {
	return rcv.LessonChatReader.ListConversationByLessons(ctx, req)
}

type ManabufV1ChatReader struct {
	SupportChatReader   *support.ChatReader
	SupportChatReaderV2 *support.ChatReader
	CoreChatReader      *core.ChatReader
	tpb.UnimplementedChatReaderServiceServer
}

func (rcv *ManabufV1ChatReader) ListConversationsInSchool(ctx context.Context, req *tpb.ListConversationsInSchoolRequest) (*tpb.ListConversationsInSchoolResponse, error) {
	return rcv.SupportChatReader.ListConversationsInSchool(ctx, req)
}

func (rcv *ManabufV1ChatReader) ListConversationsInSchoolV2(ctx context.Context, req *tpb.ListConversationsInSchoolRequest) (*tpb.ListConversationsInSchoolResponse, error) {
	return rcv.SupportChatReaderV2.ListConversationsInSchoolWithLocations(ctx, req)
}

func (rcv *ManabufV1ChatReader) ListConversationsInSchoolWithLocations(ctx context.Context, req *tpb.ListConversationsInSchoolRequest) (*tpb.ListConversationsInSchoolResponse, error) {
	return rcv.SupportChatReader.ListConversationsInSchoolWithLocations(ctx, req)
}

func (rcv *ManabufV1ChatReader) RetrieveTotalUnreadMessage(ctx context.Context, req *tpb.RetrieveTotalUnreadMessageRequest) (*tpb.RetrieveTotalUnreadMessageResponse, error) {
	return rcv.SupportChatReader.RetrieveTotalUnreadMessage(ctx, req)
}

func (rcv *ManabufV1ChatReader) RetrieveTotalUnreadConversationsWithLocations(ctx context.Context, req *tpb.RetrieveTotalUnreadConversationsWithLocationsRequest) (*tpb.RetrieveTotalUnreadConversationsWithLocationsResponse, error) {
	return rcv.SupportChatReader.RetrieveTotalUnreadConversationsWithLocations(ctx, req)
}

func (rcv *ManabufV1ChatReader) GetConversationV2(ctx context.Context, req *tpb.GetConversationV2Request) (*tpb.GetConversationV2Response, error) {
	return rcv.CoreChatReader.GetConversationV2(ctx, req)
}

type ManabufV1ChatModifier struct {
	Chat                *core.ChatServiceImpl
	SupportChatModifier *support.ChatModifier
	tpb.UnimplementedChatModifierServiceServer
}

func (rcv *ManabufV1ChatModifier) LeaveConversations(ctx context.Context, req *tpb.LeaveConversationsRequest) (*tpb.LeaveConversationsResponse, error) {
	return rcv.SupportChatModifier.LeaveConversations(ctx, req)
}

func (rcv *ManabufV1ChatModifier) JoinAllConversationsWithLocations(ctx context.Context, req *tpb.JoinAllConversationRequest) (*tpb.JoinAllConversationResponse, error) {
	return rcv.SupportChatModifier.JoinAllConversationsWithLocations(ctx, req)
}

func (rcv *ManabufV1ChatModifier) JoinAllConversations(ctx context.Context, req *tpb.JoinAllConversationRequest) (*tpb.JoinAllConversationResponse, error) {
	return rcv.SupportChatModifier.JoinAllConversations(ctx, req)
}

func (rcv *ManabufV1ChatModifier) JoinConversations(ctx context.Context, req *tpb.JoinConversationsRequest) (*tpb.JoinConversationsResponse, error) {
	return rcv.SupportChatModifier.JoinConversations(ctx, req)
}

func (rcv *ManabufV1ChatModifier) DeleteMessage(ctx context.Context, req *tpb.DeleteMessageRequest) (*tpb.DeleteMessageResponse, error) {
	return rcv.Chat.DeleteMessage(ctx, req)
}

type ManabufV1LessonChatReader struct {
	LessonChatReader *lesson.ChatReader
	tpb.UnimplementedLessonChatReaderServiceServer
}

func (rcv *ManabufV1LessonChatReader) RefreshLiveLessonSession(ctx context.Context, req *tpb.RefreshLiveLessonSessionRequest) (*tpb.RefreshLiveLessonSessionResponse, error) {
	return rcv.LessonChatReader.RefreshLiveLessonSession(ctx, req)
}
func (rcv *ManabufV1LessonChatReader) LiveLessonConversationDetail(ctx context.Context, req *tpb.LiveLessonConversationDetailRequest) (*tpb.LiveLessonConversationDetailResponse, error) {
	return rcv.LessonChatReader.LiveLessonConversationDetail(ctx, req)
}
func (rcv *ManabufV1LessonChatReader) LiveLessonConversationMessages(ctx context.Context, req *tpb.LiveLessonConversationMessagesRequest) (*tpb.LiveLessonConversationMessagesResponse, error) {
	return rcv.LessonChatReader.LiveLessonConversationMessages(ctx, req)
}
