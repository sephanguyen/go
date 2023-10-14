package mock_services

import (
	context "context"

	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/mock"
	grpc "google.golang.org/grpc"
)

type EurekaLearningObjectiveModifierServiceClient struct {
	mock.Mock
}

func (m *EurekaLearningObjectiveModifierServiceClient) UpsertLOs(arg1 context.Context, arg2 *epb.UpsertLOsRequest, arg3 ...grpc.CallOption) (*epb.UpsertLOsResponse, error) {
	args := m.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*epb.UpsertLOsResponse), args.Error(1)
}
