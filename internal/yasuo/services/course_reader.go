package services

import (
	"context"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/jackc/pgtype"
)

type CourseReaderService struct {
	DB               database.Ext
	OldCourseService *CourseService

	UserRepo interface {
		Get(context.Context, database.QueryExecer, pgtype.Text) (*entities_bob.User, error)
	}

	TeacherRepo interface {
		GetTeacherHasSchoolIDs(ctx context.Context, db database.QueryExecer, teacherID string, schoolIds []int32) (*entities_bob.Teacher, error)
		IsInSchool(ctx context.Context, db database.QueryExecer, teacherID string, schoolID int32) (bool, error)
		ManyTeacherIsInSchool(ctx context.Context, db database.QueryExecer, teacherIDs pgtype.TextArray, schoolID pgtype.Int4) (bool, error)
	}

	SchoolAdminRepo interface {
		Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities_bob.SchoolAdmin, error)
	}
}

func (s *CourseReaderService) ValidateUserSchool(ctx context.Context, req *pb.ValidateUserSchoolRequest) (*pb.ValidateUserSchoolResponse, error) {
	user, err := s.UserRepo.Get(ctx, s.DB, database.Text(req.GetUserId()))
	if err != nil {
		return nil, err
	}
	userGroup := user.Group.String

	switch userGroup {
	case entities_bob.UserGroupSchoolAdmin:
		// school admin can only upsert quiz in their school
		schoolAdmin, err := s.SchoolAdminRepo.Get(ctx, s.DB, database.Text(req.GetUserId()))
		if err != nil {
			return nil, err
		}
		if req.GetExpectSchoolId() == schoolAdmin.SchoolID.Int {
			return &pb.ValidateUserSchoolResponse{
				Result: true,
			}, nil
		}
	case entities_bob.UserGroupTeacher:
		isInSchool, err := s.TeacherRepo.IsInSchool(ctx, s.DB, req.GetUserId(), req.GetExpectSchoolId())
		if err != nil {
			return nil, err
		}
		if isInSchool {
			return &pb.ValidateUserSchoolResponse{
				Result: true,
			}, nil
		}
	case entities_bob.UserGroupAdmin:
		// admin can upsert quiz of all school
		return &pb.ValidateUserSchoolResponse{
			Result: true,
		}, nil
	}
	return &pb.ValidateUserSchoolResponse{
		Result: false,
	}, nil
}
