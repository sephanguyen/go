package services

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StudentAssignmentReaderABACService doing basic validation
type StudentAssignmentReaderABACService struct {
	*StudentAssignmentReaderService
	DB            database.Ext
	StudyPlanRepo interface {
		IsStudentAssignedItem(context.Context, database.QueryExecer,
			pgtype.Text, pgtype.TextArray) (bool, error)
	}
	GradeRepo interface {
		CheckSubmissions(ctx context.Context, db database.QueryExecer,
			gradeIDs pgtype.TextArray, studentID, status pgtype.Text) (bool, error)
	}
}

// ListSubmissions checks for request validation
func (s *StudentAssignmentReaderABACService) ListSubmissions(ctx context.Context, req *pb.ListSubmissionsRequest) (*pb.ListSubmissionsResponse, error) {
	if req.Start == nil || req.End == nil {
		return nil, status.Error(codes.InvalidArgument, "must specific Start or End")
	}

	return s.StudentAssignmentReaderService.ListSubmissions(ctx, req)
}

// RetrieveSubmissions validates if user can fetch all the requested items
func (s *StudentAssignmentReaderABACService) RetrieveSubmissions(ctx context.Context, req *pb.RetrieveSubmissionsRequest) (*pb.RetrieveSubmissionsResponse, error) {
	req.StudyPlanItemIds = golibs.Uniq(req.StudyPlanItemIds)
	userGroup := interceptors.UserGroupFromContext(ctx)
	userID := interceptors.UserIDFromContext(ctx)
	if userGroup == entities.UserGroupStudent {
		ok, err := s.StudyPlanRepo.IsStudentAssignedItem(ctx, s.DB,
			database.Text(userID), database.TextArray(req.StudyPlanItemIds))
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, status.Error(codes.PermissionDenied, "cannot fetch unassigned submissions")
		}
	}

	return s.StudentAssignmentReaderService.RetrieveSubmissions(ctx, req)
}

// RetrieveSubmissionGrades allow fetching grade result if belong to student and returned by teacher
func (s *StudentAssignmentReaderABACService) RetrieveSubmissionGrades(ctx context.Context,
	req *pb.RetrieveSubmissionGradesRequest) (*pb.RetrieveSubmissionGradesRespose, error) {
	req.SubmissionGradeIds = golibs.Uniq(req.SubmissionGradeIds)
	userGroup := interceptors.UserGroupFromContext(ctx)
	userID := interceptors.UserIDFromContext(ctx)
	if userGroup == entities.UserGroupStudent {
		ok, err := s.GradeRepo.CheckSubmissions(ctx, s.DB,
			database.TextArray(req.SubmissionGradeIds),
			database.Text(userID),
			database.Text(pb.SubmissionStatus_SUBMISSION_STATUS_RETURNED.String()))
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "cannot fetch unassigned or not returned submissions")
		}
	}

	return s.StudentAssignmentReaderService.RetrieveSubmissionGrades(ctx, req)
}

// ListSubmissions checks for request validation
func (s *StudentAssignmentReaderABACService) ListSubmissionsV2(ctx context.Context, req *pb.ListSubmissionsV2Request) (*pb.ListSubmissionsV2Response, error) {
	return s.StudentAssignmentReaderService.ListSubmissionsV2(ctx, req)
}
