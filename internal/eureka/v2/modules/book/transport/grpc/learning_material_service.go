package grpc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/book/usecase"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LearningMaterialGrpcService struct {
	updatePublishStatusLearningMaterialsUsecase usecase.UpdatePublishStatusLearningMaterials
}

func NewLearningMaterialGrpcService(
	updatePublishStatusLearningMaterialsUsecase usecase.UpdatePublishStatusLearningMaterials,
) *LearningMaterialGrpcService {
	return &LearningMaterialGrpcService{
		updatePublishStatusLearningMaterialsUsecase: updatePublishStatusLearningMaterialsUsecase,
	}
}

func validateUpdatePublishStatusLearningMaterial(e domain.LearningMaterial) error {
	if e.ID == "" {
		return fmt.Errorf("missing identity of learning material")
	}

	return nil
}

func transformLearningMaterialFromPbUpdatePublishStatusLearningMaterial(req *pb.UpdatePublishStatusLearningMaterialsRequest_PublishStatus) domain.LearningMaterial {
	lm := domain.LearningMaterial{
		ID:        req.LearningMaterialId,
		Published: req.IsPublished,
	}

	return lm
}

func (s *LearningMaterialGrpcService) UpdatePublishStatusLearningMaterials(ctx context.Context, req *pb.UpdatePublishStatusLearningMaterialsRequest) (*pb.UpdatePublishStatusLearningMaterialsResponse, error) {
	lms := make([]domain.LearningMaterial, len(req.PublishStatuses))
	res := &pb.UpdatePublishStatusLearningMaterialsResponse{}

	for idx, item := range req.PublishStatuses {
		lm := transformLearningMaterialFromPbUpdatePublishStatusLearningMaterial(item)

		if err := validateUpdatePublishStatusLearningMaterial(lm); err != nil {
			return res, status.Errorf(codes.InvalidArgument, err.Error())
		}

		lms[idx] = lm
	}

	err := s.updatePublishStatusLearningMaterialsUsecase.UpdatePublishStatusLearningMaterials(ctx, lms)

	if err != nil {
		return res, status.Errorf(codes.Internal, fmt.Errorf("learningMaterialGrpcService.UpdatePublishStatusLearningMaterials: %w", err).Error())
	}

	return res, nil
}

func (s *LearningMaterialGrpcService) ListLearningMaterialInfo(context.Context, *pb.ListLearningMaterialInfoRequest) (*pb.ListLearningMaterialInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
