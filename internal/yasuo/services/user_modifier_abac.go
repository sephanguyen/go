package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserModifierServiceABAC struct {
	*UserModifierService
}

func (s *UserModifierServiceABAC) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid params")
	}

	currentUserID := interceptors.UserIDFromContext(ctx)

	currentUGroup, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return nil, fmt.Errorf("s.UserRepo.UserGroup: %w", err)
	}

	switch currentUGroup {
	case entities.UserGroupAdmin:
		break
	case entities.UserGroupSchoolAdmin:
		schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DB, database.Text(currentUserID))
		if err != nil {
			return nil, fmt.Errorf("s.SchoolAdminRepo.Get: %w", err)
		}

		if schoolAdmin == nil {
			return nil, status.Error(codes.PermissionDenied, "cann't get school admin")
		}

		userGroup, err := s.UserRepo.UserGroup(ctx, s.DB, database.Text(req.Id))
		if err != nil {
			return nil, fmt.Errorf("s.UserRepo.UserGroup: %w", err)
		}

		switch userGroup {
		case entities.UserGroupTeacher:
			teacher, err := s.TeacherRepo.FindByID(ctx, s.DB, database.Text(req.Id))
			if err != nil {
				return nil, fmt.Errorf("s.TeacherRepo.FindByID: %w", err)
			}

			flag := 0

			for _, schoolID := range teacher.SchoolIDs.Elements {
				if schoolID.Int == schoolAdmin.SchoolID.Int {
					flag = 1
					break
				}
			}

			if flag == 0 {
				return nil, status.Error(codes.PermissionDenied, "school staff only update their teacher")
			}

			break

		case entities.UserGroupStudent:
			student, err := s.StudentRepo.Find(ctx, s.DB, database.Text(req.Id))
			if err != nil {
				return nil, fmt.Errorf("s.StudentRepo.Find: %w", err)
			}

			if student.SchoolID.Int != schoolAdmin.SchoolID.Int {
				return nil, status.Error(codes.PermissionDenied, "school staff only update their student")
			}
			break
		default:
			return nil, status.Error(codes.PermissionDenied, "only school staff can update their teacher or student")
		}
	}
	return s.UserModifierService.UpdateUserProfile(ctx, req)
}
