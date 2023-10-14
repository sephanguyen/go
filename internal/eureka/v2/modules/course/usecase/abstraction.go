package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository"
)

type CourseUsecase struct {
	ExternalCourseRepo repository.ExternalCourseRepo
	CourseBookRepo     repository.CourseBookRepo
	CourseRepo         repository.CourseRepo
}

func NewCourseUsecase(externalCourseRepo repository.ExternalCourseRepo, courseBookRepo repository.CourseBookRepo, courseRepo repository.CourseRepo) *CourseUsecase {
	return &CourseUsecase{
		ExternalCourseRepo: externalCourseRepo,
		CourseBookRepo:     courseBookRepo,
		CourseRepo:         courseRepo,
	}
}

type UpsertCourse interface {
	UpsertCourses(ctx context.Context, courses []domain.Course) error
}

type ListCourse interface {
	ListCourses(ctx context.Context, ids []string) ([]*domain.Course, error)
}
