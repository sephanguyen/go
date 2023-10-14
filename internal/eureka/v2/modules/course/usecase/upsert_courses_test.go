package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	mock_course_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/course/repository"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpsertCoursesHandler_UpsertCourses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	externalCourseRepo := &mock_course_repo.MockExternalCourseRepo{}
	courseBookRepo := &mock_course_repo.MockCourseBookRepo{}
	courseRepo := &mock_course_repo.MockCourseRepo{}

	t.Run("Upsert course successfully", func(t *testing.T) {
		// arrange
		handler := NewCourseUsecase(externalCourseRepo, courseBookRepo, courseRepo)
		courses := []domain.Course{
			{
				Name:   "course-1",
				ID:     "id-1",
				BookID: "book-1",
			},
		}

		externalCourseRepo.On("Upsert", ctx, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				actual := args[1].([]domain.Course)
				assert.Equal(t, courses[0].ID, actual[0].ID)
			}).
			Return(courses, nil)

		courseBookRepo.On("RetrieveAssociatedBook", ctx, mock.Anything).
			Once().
			Return([]*dto.CourseBookDto{
				&dto.CourseBookDto{
					BookID: pgtype.Text{String: courses[0].BookID, Status: pgtype.Present},
				},
			}, nil)

		courseBookRepo.On("Upsert", ctx, mock.Anything).Once().Return(nil)

		// act
		err := handler.UpsertCourses(ctx, courses)

		// assert
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, courseRepo)
	})

	t.Run("Upsert course failed", func(t *testing.T) {
		// arrange
		handler := NewCourseUsecase(externalCourseRepo, courseBookRepo, courseRepo)
		courses := []domain.Course{
			{
				Name: "course-3",
				ID:   "id-3",
			},
			{
				Name: "course-4",
				ID:   "id-4",
			},
		}
		externalCourseRepo.On("Upsert", ctx, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				actual := args[1].([]domain.Course)
				assert.Equal(t, courses[0].Name, actual[0].Name)
				assert.Equal(t, courses[0].ID, actual[0].ID)

				assert.Equal(t, courses[1].Name, actual[1].Name)
				assert.Equal(t, courses[1].ID, actual[1].ID)
			}).
			Return([]domain.Course{}, errors.New("some error", nil))

		// act
		err := handler.UpsertCourses(ctx, courses)

		// assert
		assert.Equal(t, err, errors.New("CourseUsecase.UpsertCourses", errors.New("some error", nil)))
		mock.AssertExpectationsForObjects(t, courseRepo)
	})
}
