package services

import (
	"context"
	"testing"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_bob_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCourseModifier_AttachMaterialsToCourse(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)

	courseRepo := &mock_bob_repositories.MockCourseRepo{}
	lessonGroupRepo := &mock_bob_repositories.MockLessonGroupRepo{}
	db := &mock_database.Ext{}

	s := &CourseModifierService{
		DB:               db,
		UnleashClientIns: mockUnleashClient,
		CourseRepo:       courseRepo,
		LessonGroupRepo:  lessonGroupRepo,
		Env:              "local",
	}
	courseId := idutil.ULIDNow()

	lessonGroupId := idutil.ULIDNow()

	testCases := []TestCase{
		{
			name: "happy case attach successfully",
			ctx:  ctx,
			req: &ypb.AttachMaterialsToCourseRequest{
				CourseId:      courseId,
				LessonGroupId: lessonGroupId,
				MaterialIds:   []string{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				courseRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&entities_bob.Course{
					ID: database.Text(courseId),
				}, nil)
				lessonGroupRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "course not found",
			ctx:  ctx,
			req: &ypb.AttachMaterialsToCourseRequest{
				CourseId:      courseId,
				LessonGroupId: lessonGroupId,
				MaterialIds:   []string{},
			},
			expectedErr: status.Error(codes.NotFound, "not found course"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabled", mock.Anything, mock.Anything).
					Return(false, nil).Once()
				courseRepo.On("FindByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				lessonGroupRepo.On("BulkUpsert", mock.Anything, mock.Anything, mock.Anything).Once().Return(status.Error(codes.NotFound, "not found course"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			_, err := s.AttachMaterialsToCourse(testCase.ctx, testCase.req.(*ypb.AttachMaterialsToCourseRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
