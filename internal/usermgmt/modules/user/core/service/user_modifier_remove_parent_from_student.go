package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	featureFlagParentLocationValidation = "User_StudentManagement_BackOffice_ParentLocationValidation"
)

func (s *UserModifierService) RemoveParentFromStudent(ctx context.Context, req *pb.RemoveParentFromStudentRequest) (*pb.RemoveParentFromStudentResponse, error) {
	shouldEnableParentLocationValidation, err := s.UnleashClient.IsFeatureEnabled(featureFlagParentLocationValidation, s.Env)
	if err != nil {
		shouldEnableParentLocationValidation = false
	}
	if err := s.validateRemoveParentRequest(req); err != nil {
		return nil, err
	}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Check if student to assign exist
		findStudentResult, err := s.StudentRepo.FindStudentProfilesByIDs(ctx, tx, database.TextArray([]string{req.StudentId}))
		if err != nil {
			return errorx.ToStatusError(status.Error(codes.Internal, err.Error()))
		}
		if len(findStudentResult) == 0 {
			return errorx.ToStatusError(status.Error(codes.InvalidArgument, fmt.Sprintf("cannot remove parents associated with un-existing student in system: %s", req.StudentId)))
		}
		// Check if parent to exist
		findParentResult, err := s.ParentRepo.GetByIds(ctx, tx, database.TextArray([]string{req.ParentId}))
		if err != nil {
			return errorx.ToStatusError(status.Error(codes.Internal, err.Error()))
		}
		if len(findParentResult) == 0 {
			return errorx.ToStatusError(status.Error(codes.InvalidArgument, fmt.Sprintf("cannot remove un-existing parents in system: %s", req.ParentId)))
		}
		// Parent should have at least one child
		if shouldEnableParentLocationValidation {
			studentParents, err := s.StudentParentRepo.FindStudentParentsByParentID(ctx, tx, req.ParentId)
			if err != nil {
				return errorx.ToStatusError(status.Error(codes.Internal, err.Error()))
			}
			if len(studentParents) <= 1 {
				return errorx.ToStatusError(status.Error(codes.InvalidArgument, fmt.Sprint(constant.InvalidRemoveParent)))
			}
		}
		err = s.StudentParentRepo.RemoveParentFromStudent(ctx, tx, database.Text(req.ParentId), database.Text(req.StudentId))
		if err != nil {
			return err
		}

		if err := s.StudentParentRepo.UpsertParentAccessPathByID(ctx, tx, []string{req.ParentId}); err != nil {
			return err
		}
		userEvents := make([]*pb.EvtUser, 0, len(findStudentResult))
		userEvents = append(userEvents, newParentRemovedFromStudentEvent(req.StudentId, req.ParentId))

		if err := s.publishUserEvent(ctx, constants.SubjectUserUpdated, userEvents...); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	response := &pb.RemoveParentFromStudentResponse{
		StudentId: req.StudentId,
		ParentId:  req.ParentId,
	}
	return response, nil
}

func (s *UserModifierService) validateRemoveParentRequest(req *pb.RemoveParentFromStudentRequest) error {
	if req.StudentId == "" {
		return errorx.ToStatusError(status.Error(codes.InvalidArgument, "student id cannot be empty"))
	}

	if req.ParentId == "" {
		return errorx.ToStatusError(status.Error(codes.InvalidArgument, "parent id cannot be empty"))
	}
	return nil
}

func newParentRemovedFromStudentEvent(studentID, parentID string) *pb.EvtUser {
	parentRemovedFromStudentEvent := &pb.EvtUser{
		Message: &pb.EvtUser_ParentRemovedFromStudent_{
			ParentRemovedFromStudent: &pb.EvtUser_ParentRemovedFromStudent{
				StudentId: studentID,
				ParentId:  parentID,
			},
		},
	}
	return parentRemovedFromStudentEvent
}
