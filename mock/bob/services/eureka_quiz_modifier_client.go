package mock_services

import (
	"context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	mock "github.com/stretchr/testify/mock"
	grpc "google.golang.org/grpc"
)

type EurekaQuizModifierServiceClient struct {
	mock.Mock
}

func (m *EurekaQuizModifierServiceClient) RemoveQuizFromLO(arg1 context.Context, arg2 *epb.RemoveQuizFromLORequest, arg3 ...grpc.CallOption) (*epb.RemoveQuizFromLOResponse, error) {
	args := m.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*epb.RemoveQuizFromLOResponse), args.Error(1)
}

func (m *EurekaQuizModifierServiceClient) DeleteQuiz(arg1 context.Context, arg2 *epb.DeleteQuizRequest, arg3 ...grpc.CallOption) (*epb.DeleteQuizResponse, error) {
	args := m.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*epb.DeleteQuizResponse), args.Error(1)
}
