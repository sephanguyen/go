package services

import (
	"context"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SubscriptionServiceABAC struct {
	*SubscriptionModifyService
}

func (s *SubscriptionServiceABAC) CreatePackage(ctx context.Context, req *pb.CreatePackageRequest) (*pb.CreatePackageResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	if (req.StartAt == nil && req.EndAt == nil) && req.Duration == 0 {
		return nil, status.Error(codes.InvalidArgument, "should have duration or (start_at, end_at)")
	}

	if req.Properties == nil {
		return nil, status.Error(codes.InvalidArgument, "properties is required")
	}

	if len(req.Properties.CanDoQuiz) == 0 &&
		len(req.Properties.CanWatchVideo) == 0 &&
		len(req.Properties.CanViewStudyGuide) == 0 {
		return nil, status.Error(codes.InvalidArgument, "should have one of properties: can_do_quiz, can_watch_video, can_view_study_guide")
	}

	startAt := req.StartAt.AsTime()
	endAt := req.EndAt.AsTime()

	if startAt.After(endAt) {
		return nil, status.Error(codes.InvalidArgument, "startAt must be before endAt")
	}

	if time.Now().After(endAt) {
		return nil, status.Error(codes.InvalidArgument, "endAt must be in the future")
	}

	return s.SubscriptionModifyService.CreatePackage(ctx, req)
}

func (s *SubscriptionServiceABAC) ToggleActivePackage(ctx context.Context, req *pb.ToggleActivePackageRequest) (*pb.ToggleActivePackageResponse, error) {
	if req.PackageId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing packageId")
	}

	return s.SubscriptionModifyService.ToggleActivePackage(ctx, req)
}

func (s *SubscriptionServiceABAC) AddStudentPackage(ctx context.Context, req *pb.AddStudentPackageRequest) (*pb.AddStudentPackageResponse, error) {
	if req.PackageId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing packageId")
	}

	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing studentId")
	}

	return s.SubscriptionModifyService.AddStudentPackage(ctx, req)
}

func (s *SubscriptionServiceABAC) ToggleActiveStudentPackage(ctx context.Context, req *pb.ToggleActiveStudentPackageRequest) (*pb.ToggleActiveStudentPackageResponse, error) {
	if req.StudentPackageId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing student packageId")
	}

	return s.SubscriptionModifyService.ToggleActiveStudentPackage(ctx, req)
}

func (s *SubscriptionServiceABAC) AddStudentPackageCourse(ctx context.Context, req *pb.AddStudentPackageCourseRequest) (*pb.AddStudentPackageCourseResponse, error) {
	if len(req.CourseIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing courseId")
	}

	if req.StudentId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing studentId")
	}

	if req.StartAt.CheckValid() != nil || req.EndAt.CheckValid() != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid time")
	}

	startAt := req.StartAt.AsTime()
	endAt := req.EndAt.AsTime()

	if startAt.After(endAt) {
		return nil, status.Error(codes.InvalidArgument, "startAt must be before endAt")
	}

	return s.SubscriptionModifyService.AddStudentPackageCourse(ctx, req)
}

func (s *SubscriptionServiceABAC) WrapperRegisterStudentClass(ctx context.Context, req *pb.WrapperRegisterStudentClassRequest) (*pb.WrapperRegisterStudentClassResponse, error) {
	if err := s.SubscriptionModifyService.WrapperRegisterStudentClass(ctx, req); err != nil {
		return nil, err
	}
	return &pb.WrapperRegisterStudentClassResponse{Successful: true}, nil
}

func (s *SubscriptionServiceABAC) RegisterStudentClass(ctx context.Context, req *pb.RegisterStudentClassRequest) (*pb.RegisterStudentClassResponse, error) {
	if err := validateRegisterStudentClassRequest(req); err != nil {
		return nil, err
	}
	if err := s.SubscriptionModifyService.RegisterStudentClass(ctx, req); err != nil {
		return nil, err
	}
	return &pb.RegisterStudentClassResponse{Successful: true}, nil
}

func validateRegisterStudentClassRequest(req *pb.RegisterStudentClassRequest) error {
	if len(req.ClassesInformation) == 0 {
		return status.Error(codes.InvalidArgument, "missing classes information")
	}
	for _, classInformation := range req.ClassesInformation {
		switch {
		case classInformation.StudentPackageId == "":
			return status.Error(codes.InvalidArgument, "student_package_id cannot be empty")
		case classInformation.StudentId == "":
			return status.Error(codes.InvalidArgument, "student_id cannot be empty")
		case classInformation.ClassId == "":
			return status.Error(codes.InvalidArgument, "class_id cannot be empty")
		case !classInformation.StartTime.IsValid():
			return status.Error(codes.InvalidArgument, "start_time invalid")
		case !classInformation.EndTime.IsValid():
			return status.Error(codes.InvalidArgument, "end_time invalid")
		}
	}
	return nil
}

func (s *SubscriptionServiceABAC) EditTimeStudentPackage(ctx context.Context, req *pb.EditTimeStudentPackageRequest) (*pb.EditTimeStudentPackageResponse, error) {
	if req.StudentPackageId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing student packageId")
	}

	if req.StartAt.CheckValid() != nil || req.EndAt.CheckValid() != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid time")
	}

	startAt := req.StartAt.AsTime()
	endAt := req.EndAt.AsTime()

	if startAt.After(endAt) {
		return nil, status.Error(codes.InvalidArgument, "startAt must be before endAt")
	}

	return s.SubscriptionModifyService.EditTimeStudentPackage(ctx, req)
}

func (s *SubscriptionServiceABAC) RetrieveStudentPackagesUnderCourse(ctx context.Context, req *pb.RetrieveStudentPackagesUnderCourseRequest) (*pb.RetrieveStudentPackagesUnderCourseResponse, error) {
	return s.SubscriptionModifyService.RetrieveStudentPackagesUnderCourse(ctx, req)
}
