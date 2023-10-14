package grpc

import (
	"context"

	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentStudyPlanService struct {
}

var _ pb.StudentStudyPlanServiceServer = (*StudentStudyPlanService)(nil)

func (a *StudentStudyPlanService) ListStudentStudyPlan(context.Context, *pb.ListStudentStudyPlanRequest) (*pb.ListStudentStudyPlanResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (a *StudentStudyPlanService) GetStudentStudyPlanStatus(context.Context, *pb.GetStudentStudyPlanStatusRequest) (*pb.GetStudentStudyPlanStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (a *StudentStudyPlanService) ListStudentStudyPlanItem(context.Context, *pb.ListStudentStudyPlanItemRequest) (*pb.ListStudentStudyPlanItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
