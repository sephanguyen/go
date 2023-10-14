package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/repositories"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CourseService struct {
	CourseAccessPathRepo interface {
		GetCourseAccessPathByUCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (mapCourseAccess map[string]interface{}, err error)
	}
}

func (s *CourseService) GetMapLocationAccessCourseForCourseIDs(ctx context.Context, db database.QueryExecer, courseIDs []string) (mapLocationAccessCourse map[string]interface{}, err error) {
	mapLocationAccessCourse, err = s.CourseAccessPathRepo.GetCourseAccessPathByUCourseIDs(ctx, db, courseIDs)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't get map course access location")
	}
	return
}

func NewCourseService() *CourseService {
	return &CourseService{
		CourseAccessPathRepo: &repositories.CourseAccessPathRepo{},
	}
}
