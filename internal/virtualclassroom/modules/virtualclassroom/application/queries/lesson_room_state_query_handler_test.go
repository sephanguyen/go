package queries

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	virDomain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"

	"github.com/golang-migrate/migrate/v4/database/pgx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLessonRoomStateQuery_GetLessonRoomStateByLessonID(t *testing.T) {
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	lessonRoomStateRepo := &mock_virtual_repo.MockLessonRoomStateRepo{}
	lessonId := "lesson-1"
	now := time.Now()
	state := &domain.LessonRoomState{
		LessonID:        lessonId,
		SpotlightedUser: "user-1",
		WhiteboardZoomState: &virDomain.WhiteboardZoomState{
			PdfScaleRatio: 23.32,
			CenterX:       243.5,
			CenterY:       -432.034,
			PdfWidth:      234.43,
			PdfHeight:     -0.33424,
		},
		Recording: &virDomain.CompositeRecordingState{
			ResourceID:  "resource-id",
			SID:         "s-id",
			UID:         123342,
			IsRecording: true,
			Creator:     "user-id-1",
		},
		SessionTime: &now,
	}

	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		result   *domain.LessonRoomState
		hasError bool
	}{
		{
			name: "success with existed result",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, lessonId).
					Return(state, nil).Once()
			},
			result: state,
		},
		{
			name: "success with error not found",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, lessonId).
					Return(nil, domain.ErrLessonRoomStateNotFound).Once()
			},
			result: &domain.LessonRoomState{
				Recording: &virDomain.CompositeRecordingState{},
			},
		},
		{
			name: "error",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRoomStateRepo.On("GetLessonRoomStateByLessonID", ctx, db, lessonId).
					Return(nil, pgx.ErrDatabaseDirty).Once()
			},
			hasError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			q := &LessonRoomStateQuery{
				WrapperDBConnection: wrapperConnection,
				LessonRoomStateRepo: lessonRoomStateRepo,
			}
			res, err := q.GetLessonRoomStateByLessonID(ctx, LessonRoomStateQueryPayload{LessonID: lessonId})

			if tc.hasError {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, res, tc.result)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}

}
