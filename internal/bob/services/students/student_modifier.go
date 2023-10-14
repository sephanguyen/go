package services

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type StudentModifierServices struct {
	bpb.UnimplementedStudentModifierServiceServer
	DB                     database.Ext
	UserMgmtStudentService services.UserMgmtStudentSvc
}

func NewStudentModifierServices(
	db database.Ext,
	userMgmtStudentService services.UserMgmtStudentSvc,
) *StudentModifierServices {
	return &StudentModifierServices{
		DB:                     db,
		UserMgmtStudentService: userMgmtStudentService,
	}
}

func (s *StudentModifierServices) DeleteStudentComments(ctx context.Context, req *bpb.DeleteStudentCommentsRequest) (*bpb.DeleteStudentCommentsResponse, error) {
	if req.CommentIds == nil {
		return nil, status.Error(codes.InvalidArgument, "comment ids have to not empty")
	}
	headers, ok := metadata.FromIncomingContext(ctx)
	var pkg, token, version string
	if ok {
		pkg = headers["pkg"][0]
		token = headers["token"][0]
		version = headers["version"][0]
	}

	userMgmtResponse, err := s.UserMgmtStudentService.DeleteStudentComments(
		metadata.AppendToOutgoingContext(ctx, "pkg", pkg, "version", version, "token", token),
		&upb.DeleteStudentCommentsRequest{
			CommentIds: req.CommentIds,
		})
	if err != nil {
		return nil, err
	}
	return &bpb.DeleteStudentCommentsResponse{
		Successful: userMgmtResponse.Successful,
	}, nil
}

// func (s *StudentModifierServices) UpdateProfile(context.Context, *bpb.UpdateProfileRequest) (*bpb.UpdateProfileResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method UpdateProfile not implemented")
// }

// func (s *StudentModifierServices) CreateStudentEventLogs(context.Context, *bpb.CreateStudentEventLogsRequest) (*bpb.CreateStudentEventLogsResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method CreateStudentEventLogs not implemented")
// }
