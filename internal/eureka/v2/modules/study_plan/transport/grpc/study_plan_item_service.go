package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/usecase"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

type StudyPlanItemService struct {
	StudyPlanItemUseCase *usecase.StudyPlanItemUseCase
}

func NewStudyPlanItemService(studyPlanItemUseCase *usecase.StudyPlanItemUseCase) *StudyPlanItemService {
	return &StudyPlanItemService{
		StudyPlanItemUseCase: studyPlanItemUseCase,
	}
}

var _ pb.StudyPlanItemServiceServer = (*StudyPlanItemService)(nil)

func (b *StudyPlanItemService) UpsertStudyPlanItem(ctx context.Context, req *pb.UpsertStudyPlanItemRequest) (*pb.UpsertStudyPlanItemResponse, error) {
	studyPlanItem := domain.StudyPlanItem{
		StudyPlanItemID: req.StudyPlanItemId,
		StudyPlanID:     req.StudyPlanId,
		Name:            req.Name,
		StartDate:       req.StartDate.AsTime(),
		EndDate:         req.EndDate.AsTime(),
		DisplayOrder:    int(req.DisplayOrder),
		Status:          req.Status.String(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		DeletedAt:       nil,
		LmList:          req.LmIds,
	}
	if studyPlanItem.StudyPlanItemID == "" {
		studyPlanItem.StudyPlanItemID = idutil.ULIDNow()
	}
	err := b.StudyPlanItemUseCase.UpsertStudyPlanItems(ctx, []domain.StudyPlanItem{studyPlanItem})
	if err != nil {
		return nil, fmt.Errorf("UpsertStudyPlanItems: %w", err)
	}
	return &pb.UpsertStudyPlanItemResponse{}, nil
}
