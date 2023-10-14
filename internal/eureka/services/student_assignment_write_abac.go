package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StudentAssignmentWriteABACService is abac wrapper
type StudentAssignmentWriteABACService struct {
	*StudentAssignmentWriteService
	DB             database.Ext
	AssignmentRepo interface {
		IsStudentAssigned(ctx context.Context, db database.QueryExecer, studyPlanItemID, assignmentID, studentID pgtype.Text) (bool, error)
	}
}

// SubmitAssignment check if student was assigned to assignment before calling next
func (s *StudentAssignmentWriteABACService) SubmitAssignment(ctx context.Context, req *pb.SubmitAssignmentRequest) (*pb.SubmitAssignmentResponse, error) {
	if req.Submission == nil {
		return nil, status.Error(codes.InvalidArgument, codes.InvalidArgument.String())
	}

	groupID := interceptors.UserGroupFromContext(ctx)

	studentID := ""

	switch groupID {
	case cpb.UserGroup_USER_GROUP_TEACHER.String(), constant.RoleTeacher:
		studentID = req.GetSubmission().GetStudentId()
	default:
		studentID = interceptors.UserIDFromContext(ctx)
	}

	ok, err := s.AssignmentRepo.IsStudentAssigned(ctx, s.DB,
		database.Text(req.Submission.StudyPlanItemId),
		database.Text(req.Submission.AssignmentId),
		database.Text(studentID))
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, status.Error(codes.PermissionDenied, "non-assigned assignment")
	}

	return s.StudentAssignmentWriteService.SubmitAssignment(ctx, req)
}
