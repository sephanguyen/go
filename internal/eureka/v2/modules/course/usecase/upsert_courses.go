package usecase

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
)

func (c *CourseUsecase) UpsertCourses(ctx context.Context, courses []domain.Course) error {
	returnCourses, err := c.ExternalCourseRepo.Upsert(ctx, courses)
	if err != nil {
		return errors.New("CourseUsecase.UpsertCourses", err)
	}

	if len(returnCourses) > 0 {
		var courseBooks []*dto.CourseBookDto
		for _, course := range returnCourses {
			if strings.TrimSpace(course.BookID) != "" {
				// validate Books that are not yet associated to any course
				associatedBooks, err := c.CourseBookRepo.RetrieveAssociatedBook(ctx, course.BookID)
				if err != nil {
					return errors.New("CourseUsecase.UpsertCourses", err)
				}
				err = validateAssociatedBook(associatedBooks)
				if err != nil {
					return err
				}

				courseBook, er := dto.NewCourseBookDtoFromEntity(course.ID, course.BookID)
				if er == nil {
					courseBooks = append(courseBooks, courseBook)
				}
			}
		}

		if len(courseBooks) > 0 {
			err := c.CourseBookRepo.Upsert(ctx, courseBooks)
			if err != nil {
				return errors.New("CourseUsecase.UpsertCourseBooks", err)
			}
		}
	}
	return nil
}

func validateAssociatedBook(cbs []*dto.CourseBookDto) error {
	if len(cbs) == 0 {
		return errors.NewConversionError("book not found", nil)
	}
	return nil
}
