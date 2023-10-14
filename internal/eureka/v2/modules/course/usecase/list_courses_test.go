package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/repository/postgres/dto"
	mock_course_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/course/repository"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCoursesHandler_ListCourses(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	externalCourseRepo := &mock_course_repo.MockExternalCourseRepo{}
	courseBookRepo := &mock_course_repo.MockCourseBookRepo{}
	courseRepo := &mock_course_repo.MockCourseRepo{}

	t.Run("List course successfully", func(t *testing.T) {
		// arrange
		handler := NewCourseUsecase(externalCourseRepo, courseBookRepo, courseRepo)
		courseIDs := []string{"course-id-1", "course-id-2"}
		courseDtos := []*dto.CourseDto{
			{
				ID: pgtype.Text{String: "course-id-1", Status: pgtype.Present},
			},
			{
				ID: pgtype.Text{String: "course-id-2", Status: pgtype.Present},
			},
		}
		courseRepo.On("RetrieveByIDs", ctx, mock.Anything).
			Once().
			Run(func(args mock.Arguments) {
				actual := args[1].([]string)
				assert.Equal(t, courseIDs[0], actual[0])
				assert.Equal(t, courseIDs[1], actual[1])
			}).
			Return(courseDtos, nil)

		// act
		courses, err := handler.ListCourses(ctx, courseIDs)

		// assert
		assert.Nil(t, err)
		assert.NotNil(t, courses)
		mock.AssertExpectationsForObjects(t, externalCourseRepo)
	})
}
