package grpc

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	mock_usecase "github.com/manabie-com/backend/mock/eureka/v2/modules/book/usecase/repo"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestLearningMaterialService_UpdatedPublishStatusLearningMaterials(t *testing.T) {
	ctx := context.Background()
	lmUseCase := &mock_usecase.MockLearningMaterialUsecase{}

	svc := NewLearningMaterialGrpcService(lmUseCase)

	testCases := []TestCase{{
		name: "happy case",
		req: &pb.UpdatePublishStatusLearningMaterialsRequest{
			PublishStatuses: []*pb.UpdatePublishStatusLearningMaterialsRequest_PublishStatus{
				{
					LearningMaterialId: "A_1",
					IsPublished:        false,
				},
				{
					LearningMaterialId: "A_2",
					IsPublished:        true,
				},
			},
		},
		setup: func(ctx context.Context) {
			lmUseCase.On("UpdatePublishStatusLearningMaterials", mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
				assert.Equal(t, args[1], []domain.LearningMaterial{{
					ID:        "A_1",
					Published: false,
				}, {
					ID:        "A_2",
					Published: true,
				}})

			}).Return(nil)
		},
		expectedResp: &pb.UpdatePublishStatusLearningMaterialsResponse{},
	}, {
		name: "Validate: missing learning material id",
		req: &pb.UpdatePublishStatusLearningMaterialsRequest{
			PublishStatuses: []*pb.UpdatePublishStatusLearningMaterialsRequest_PublishStatus{
				{
					LearningMaterialId: "A_1",
					IsPublished:        false,
				},
				{
					IsPublished: true,
				},
			},
		},
		setup:       func(ctx context.Context) {},
		expectedErr: status.Errorf(codes.InvalidArgument, fmt.Errorf("missing identity of learning material").Error()),
	}, {
		name: "UseCase: UpdatePublishStatusLearningMaterials error",
		req: &pb.UpdatePublishStatusLearningMaterialsRequest{
			PublishStatuses: []*pb.UpdatePublishStatusLearningMaterialsRequest_PublishStatus{
				{
					LearningMaterialId: "A_1",
					IsPublished:        false,
				},
				{
					LearningMaterialId: "A_2",
					IsPublished:        true,
				},
			},
		},
		setup: func(ctx context.Context) {
			lmUseCase.On("UpdatePublishStatusLearningMaterials", mock.Anything, mock.Anything).Once().Return(puddle.ErrNotAvailable)
		},
		expectedErr: status.Errorf(codes.Internal, fmt.Errorf("learningMaterialGrpcService.UpdatePublishStatusLearningMaterials: %w", puddle.ErrNotAvailable).Error()),
	}}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(ctx)

			resp, err := svc.UpdatePublishStatusLearningMaterials(ctx, testCase.req.(*pb.UpdatePublishStatusLearningMaterialsRequest))

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr, err)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
