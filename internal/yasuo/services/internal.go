package services

import (
	"context"

	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewInternalService() *InternalService {
	return &InternalService{}
}

type InternalService struct{}

// Deprecated: Already move to notificationmgmt
func (rcv *InternalService) RetrievePushedNotificationMessages(ctx context.Context, req *ypb.RetrievePushedNotificationMessageRequest) (*ypb.RetrievePushedNotificationMessageResponse, error) {
	return &ypb.RetrievePushedNotificationMessageResponse{}, status.Error(codes.Unimplemented, "Deprecated: Already move to notificationmgmt")
}
