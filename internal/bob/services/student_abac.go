package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StudentServiceABAC struct {
	*StudentService
}

func (s *StudentServiceABAC) GetStudentProfile(ctx context.Context, req *pb.GetStudentProfileRequest) (*pb.GetStudentProfileResponse, error) {
	if len(req.StudentIds) == 0 {
		req.StudentIds = []string{interceptors.UserIDFromContext(ctx)}
	}

	if n := len(req.StudentIds); n > 200 {
		return nil, status.Error(codes.InvalidArgument, "number of ID in validStudentIDsrequest must be less than 200")
	}

	return s.StudentService.GetStudentProfile(ctx, req)
}

func (s *StudentServiceABAC) RetrieveLearningProgress(ctx context.Context, req *pb.RetrieveLearningProgressRequest) (*pb.RetrieveLearningProgressResponse, error) {
	if req.From == nil || req.To == nil {
		return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	if !dateRangeValid(req.From, req.To) {
		return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	perm := canProcessStudentData(ctx, req.StudentId)

	if !perm {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return s.StudentService.RetrieveLearningProgress(ctx, req)
}

func (s *StudentServiceABAC) AssignPresetStudyPlans(ctx context.Context, req *pb.AssignPresetStudyPlansRequest) (*pb.AssignPresetStudyPlansResponse, error) {
	return s.StudentService.AssignPresetStudyPlans(ctx, req)
}

func (s *StudentServiceABAC) RetrieveStudentStudyPlans(ctx context.Context, req *pb.RetrieveStudentStudyPlansRequest) (*pb.RetrieveStudentStudyPlansResponse, error) {
	if !dateRangeValid(req.From, req.To) {
		return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	perm := canProcessStudentData(ctx, req.StudentId)

	if !perm {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return s.StudentService.RetrieveStudentStudyPlans(ctx, req)
}

func (s *StudentServiceABAC) RetrieveStudentStudyPlanWeeklies(ctx context.Context, req *pb.RetrieveStudentStudyPlanWeekliesRequest) (*pb.RetrieveStudentStudyPlanWeekliesResponse, error) {
	if !dateRangeValid(req.From, req.To) {
		return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	perm := canProcessStudentData(ctx, req.StudentId)
	if !perm {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return s.StudentService.RetrieveStudentStudyPlanWeeklies(ctx, req)
}

func (s *StudentServiceABAC) RetrieveDailyLOFinished(ctx context.Context, req *pb.RetrieveDailyLOFinishedRequest) (*pb.RetrieveDailyLOFinishedResponse, error) {
	return s.StudentService.RetrieveDailyLOFinished(ctx, req)
}

func (s *StudentServiceABAC) CountTotalLOsFinished(ctx context.Context, req *pb.CountTotalLOsFinishedRequest) (*pb.CountTotalLOsFinishedResponse, error) {
	if !dateRangeValid(req.From, req.To) {
		return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	perm := canProcessStudentData(ctx, req.StudentId)

	if !perm {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return s.StudentService.CountTotalLOsFinished(ctx, req)
}

func (s *StudentServiceABAC) RetrieveOverdueTopic(ctx context.Context, req *pb.RetrieveOverdueTopicRequest) (*pb.RetrieveOverdueTopicResponse, error) {
	perm := canProcessStudentData(ctx, req.StudentId)

	if !perm {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return s.StudentService.RetrieveOverdueTopic(ctx, req)
}

func (s *StudentServiceABAC) RetrieveCompletedTopicWeeklies(ctx context.Context, req *pb.RetrieveCompletedTopicWeekliesRequest) (*pb.RetrieveCompletedTopicWeekliesResponse, error) {
	perm := canProcessStudentData(ctx, req.StudentId)

	if !perm {
		return nil, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	return s.StudentService.RetrieveCompletedTopicWeeklies(ctx, req)
}
