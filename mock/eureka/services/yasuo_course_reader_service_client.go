package mock_services

import (
	"context"

	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/stretchr/testify/mock"
	grpc "google.golang.org/grpc"
)

type YasuoCourseReaderServiceClient struct {
	mock.Mock
}

func (m *YasuoCourseReaderServiceClient) ValidateUserSchool(arg1 context.Context, arg2 *ypb.ValidateUserSchoolRequest, arg3 ...grpc.CallOption) (*ypb.ValidateUserSchoolResponse, error) {
	args := m.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ypb.ValidateUserSchoolResponse), args.Error(1)
}
