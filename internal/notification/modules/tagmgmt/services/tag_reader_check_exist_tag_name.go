package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (rcv *TagMgmtReaderService) CheckExistTagName(ctx context.Context, req *npb.CheckExistTagNameRequest) (*npb.CheckExistTagNameResponse, error) {
	if req.TagName == "" {
		return nil, status.Error(codes.InvalidArgument, "TagName is empty")
	}
	isExist, err := rcv.TagRepo.DoesTagNameExist(ctx, rcv.DB, database.Text(req.TagName))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CheckExistTagName: %v", err))
	}
	return &npb.CheckExistTagNameResponse{IsExist: isExist}, nil
}
