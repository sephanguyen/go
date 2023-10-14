package clients

import (
	"context"

	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"google.golang.org/grpc"
)

type ConversationClient struct {
	conversationClient cpb.ConversationModifierServiceClient
}

type ConversationClientInterface interface {
	CreateConversation(ctx context.Context, req *cpb.CreateConversationRequest) (*cpb.CreateConversationResponse, error)
}

func InitConversationClient(connect *grpc.ClientConn) *ConversationClient {
	conversationClient := cpb.NewConversationModifierServiceClient(connect)
	return &ConversationClient{
		conversationClient: conversationClient,
	}
}

func CreateRequestCreateConversation(conversationName string, memberIDs []string) *cpb.CreateConversationRequest {
	return &cpb.CreateConversationRequest{Name: conversationName, MemberIds: memberIDs}
}

func (c *ConversationClient) CreateConversation(ctx context.Context, req *cpb.CreateConversationRequest) (*cpb.CreateConversationResponse, error) {
	return c.conversationClient.CreateConversation(ctx, req)
}
