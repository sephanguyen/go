package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/transport"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/study_plan/usecase"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"
)

type StudyPlanService struct {
	StudyPlanUsecase usecase.StudyPlanUsecase
}

func NewStudyPlanService(studyPlanUsecase usecase.StudyPlanUsecase) *StudyPlanService {
	return &StudyPlanService{
		StudyPlanUsecase: studyPlanUsecase,
	}
}

var _ pb.StudyPlanServiceServer = (*StudyPlanService)(nil)

func (a *StudyPlanService) UpsertStudyPlan(ctx context.Context, req *pb.UpsertStudyPlanRequest) (*pb.UpsertStudyPlanResponse, error) {
	if err := a.validateUpsertStudyPlanRequest(req); err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	studyPlan := a.transformStudyPlanPb(req)

	studyPlanID, err := a.StudyPlanUsecase.UpsertStudyPlan(ctx, studyPlan)
	if err != nil {
		return nil, errors.NewGrpcError(err, transport.GrpcErrorMap)
	}

	return &pb.UpsertStudyPlanResponse{
		StudyPlanId: studyPlanID,
	}, nil
}

func (a *StudyPlanService) validateUpsertStudyPlanRequest(req *pb.UpsertStudyPlanRequest) error {
	// TO-DO: validate course_id, acadaemic_year in database
	if req.CourseId == "" {
		return errors.NewValidationError("req must have course id", nil)
	}
	if req.AcademicYear == "" {
		return errors.NewValidationError("req must have academic year", nil)
	}
	if req.Name == "" {
		return errors.NewValidationError("req must have name", nil)
	}

	return nil
}

func (a *StudyPlanService) transformStudyPlanPb(studyPlan *pb.UpsertStudyPlanRequest) domain.StudyPlan {
	return domain.StudyPlan{
		Name:         studyPlan.Name,
		CourseID:     studyPlan.CourseId,
		AcademicYear: studyPlan.AcademicYear,
		Status:       domain.StudyPlanStatus(studyPlan.Status.String()),
	}
}
