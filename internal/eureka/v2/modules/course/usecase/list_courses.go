package usecase

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
)

func (c *CourseUsecase) ListCourses(ctx context.Context, ids []string) ([]*domain.Course, error) {
	courseDtos, err := c.CourseRepo.RetrieveByIDs(ctx, ids)
	if err != nil {
		return nil, errors.New("CourseUsecase.ListCourses", err)
	}
	if len(courseDtos) > 0 {
		var courses []*domain.Course
		for _, courseDto := range courseDtos {
			courses = append(courses, courseDto.ToCourseEntity())
		}
		return courses, nil
	}
	return nil, nil
}
