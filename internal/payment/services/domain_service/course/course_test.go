package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCourseService_CheckCourseExist(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	t.Run("Happy case", func(t *testing.T) {
		courseRepo := &mockRepositories.MockCourseAccessPathRepo{}
		courseRepo.On("GetCourseAccessPathByUCourseIDs", ctx, mock.Anything, mock.Anything).Return(map[string]interface{}{}, nil)
		courseService := CourseService{CourseAccessPathRepo: courseRepo}
		db := new(mockDb.Ext)
		_, err := courseService.GetMapLocationAccessCourseForCourseIDs(ctx, db, []string{"1"})
		require.Nil(t, err)
		mock.AssertExpectationsForObjects(t, db, courseRepo)
	})

	t.Run("Error when get from repo", func(t *testing.T) {
		courseRepo := &mockRepositories.MockCourseAccessPathRepo{}
		courseRepo.On("GetCourseAccessPathByUCourseIDs", ctx, mock.Anything, mock.Anything).Return(map[string]interface{}{}, constant.ErrDefault)
		courseService := CourseService{CourseAccessPathRepo: courseRepo}
		db := new(mockDb.Ext)
		_, err := courseService.GetMapLocationAccessCourseForCourseIDs(ctx, db, []string{"1"})
		require.NotNil(t, err)
		mock.AssertExpectationsForObjects(t, db, courseRepo)
	})

}
