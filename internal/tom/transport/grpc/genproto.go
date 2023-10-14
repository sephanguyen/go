package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/tom/app/core"
	"github.com/manabie-com/backend/internal/tom/app/support"
	chat "github.com/manabie-com/backend/internal/tom/infra/chat"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
)

type GenprotoChatService struct {
	ChatReader        *core.ChatReader
	Chat              *core.ChatServiceImpl
	ChatInfra         *chat.Server
	SupportChatReader *support.ChatReader
	pb.UnimplementedChatServiceServer
}

func (rcv *GenprotoChatService) GetConversation(ctx context.Context, req *pb.GetConversationRequest) (*pb.GetConversationResponse, error) {
	return rcv.ChatReader.GetConversation(ctx, req)
}

func (rcv *GenprotoChatService) ConversationDetail(ctx context.Context, req *pb.ConversationDetailRequest) (*pb.ConversationDetailResponse, error) {
	return rcv.ChatReader.ConversationDetail(ctx, req)
}

func (rcv *GenprotoChatService) ConversationList(ctx context.Context, req *pb.ConversationListRequest) (*pb.ConversationListResponse, error) {
	return rcv.SupportChatReader.ConversationList(ctx, req)
}

func (rcv *GenprotoChatService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	return rcv.Chat.SendMessage(ctx, req)
}

func (rcv *GenprotoChatService) SeenMessage(ctx context.Context, req *pb.SeenMessageRequest) (*pb.SeenMessageResponse, error) {
	return rcv.Chat.SeenMessage(ctx, req)
}

func (rcv *GenprotoChatService) RetrievePushedNotificationMessages(ctx context.Context, req *pb.RetrievePushedNotificationMessageRequest) (*pb.RetrievePushedNotificationMessageResponse, error) {
	return rcv.Chat.RetrievePushedNotificationMessages(ctx, req)
}

func (rcv *GenprotoChatService) SubscribeV2(req *pb.SubscribeV2Request, srv pb.ChatService_SubscribeV2Server) error {
	return rcv.ChatInfra.SubscribeV2(req, srv)
}

func (rcv *GenprotoChatService) PingSubscribeV2(ctx context.Context, req *pb.PingSubscribeV2Request) (*pb.PingSubscribeV2Response, error) {
	return rcv.ChatInfra.PingSubscribeV2(ctx, req)
}
