package external

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/course/domain"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_course_repo "github.com/manabie-com/backend/mock/eureka/v2/modules/course/repository"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCourseRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	ctx = interceptors.NewIncomingContext(ctx)
	defer cancel()
	mockDB := &mock_database.Ext{}
	courseModifierClient := &mock_course_repo.MockMasterDataCourseClient{}
	courseRepo := &CourseRepo{CourseClient: courseModifierClient}
	validCourseReq := []domain.Course{
		{
			ID:   "course-id-1",
			Name: "course-name-1",
		},
		{
			ID:   "course-id-2",
			Name: "course-name-2",
		},
	}

	t.Run("successfully", func(t *testing.T) {
		batchResults := &mock_database.BatchResults{}
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Exec").Once().Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)

		courseModifierClient.On("UpsertCourses", mock.Anything, mock.Anything).Once().Return(&mpb.UpsertCoursesResponse{Successful: true}, nil)

		_, err := courseRepo.Upsert(ctx, validCourseReq)
		require.Nil(t, err)
	})
}
