package controller

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVirtualClassRoomLog_LogWhenAttendeeJoin(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	mock_repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name          string
		lessonID      string
		attendeeID    string
		setup         func(ctx context.Context)
		createdNewLog bool
		hasError      bool
	}{
		{
			name:       "new log successfully when these are no any current log",
			lessonID:   "lesson-id-1",
			attendeeID: "user-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On("GetLatestByLessonID", ctx, db, "lesson-id-1").
					Return(nil, pgx.ErrNoRows).Once()
				mock_repo.On("Create", ctx, db, mock.Anything).Run(func(args mock.Arguments) {
					log := args[2].(*repo.VirtualClassRoomLogDTO)
					assert.NotEmpty(t, log.LogID.String)
					assert.Equal(t, "lesson-id-1", log.LessonID.String)
					assert.False(t, log.IsCompleted.Bool)
					assert.ElementsMatch(t, database.FromTextArray(log.AttendeeIDs), []string{"user-id-1"})
				}).Return(nil).Once()
			},
			createdNewLog: true,
			hasError:      false,
		},
		{
			name:       "get latest log fail",
			lessonID:   "lesson-id-1",
			attendeeID: "user-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On("GetLatestByLessonID", ctx, db, "lesson-id-1").
					Return(nil, errors.New("error")).Once()
			},
			hasError: true,
		},
		{
			name:       "new log successfully when current log completed",
			lessonID:   "lesson-id-1",
			attendeeID: "user-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On("GetLatestByLessonID", ctx, db, "lesson-id-1").
					Return(&repo.VirtualClassRoomLogDTO{
						LogID:       database.Text("log-id-1"),
						LessonID:    database.Text("lesson-id-1"),
						IsCompleted: database.Bool(true),
					}, nil).Once()
				mock_repo.On("Create", ctx, db, mock.Anything).Run(func(args mock.Arguments) {
					log := args[2].(*repo.VirtualClassRoomLogDTO)
					assert.NotEmpty(t, log.LogID.String)
					assert.Equal(t, "lesson-id-1", log.LessonID.String)
					assert.False(t, log.IsCompleted.Bool)
					assert.ElementsMatch(t, database.FromTextArray(log.AttendeeIDs), []string{"user-id-1"})
				}).Return(nil).Once()
			},
			createdNewLog: true,
			hasError:      false,
		},
		{
			name:       "update log successfully",
			lessonID:   "lesson-id-1",
			attendeeID: "user-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On("GetLatestByLessonID", ctx, db, "lesson-id-1").
					Return(&repo.VirtualClassRoomLogDTO{
						LogID:       database.Text("log-id-1"),
						LessonID:    database.Text("lesson-id-1"),
						IsCompleted: database.Bool(false),
					}, nil).Once()
				mock_repo.On("AddAttendeeIDByLessonID", ctx, db, "lesson-id-1", "user-id-1").
					Return(nil).Once()
			},
			createdNewLog: false,
			hasError:      false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)
			v := &VirtualClassRoomLogService{
				WrapperConnection: wrapperConnection,
				Repo:              mock_repo,
			}
			createdNewLogActual, err := v.LogWhenAttendeeJoin(ctx, tc.lessonID, tc.attendeeID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.createdNewLog, createdNewLogActual)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				mock_repo,
				mockUnleashClient,
			)
		})
	}
}

func TestVirtualClassRoomLogService_LogWhenGetRoomState(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	mock_repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID string
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:     "success case",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-id-1",
					repo.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
			},
			hasError: false,
		},
		{
			name:     "failed case",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-id-1",
					repo.TotalTimesGettingRoomState,
				).
					Return(errors.New("error")).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)
			v := &VirtualClassRoomLogService{
				WrapperConnection: wrapperConnection,
				Repo:              mock_repo,
			}
			err := v.LogWhenGetRoomState(ctx, tc.lessonID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				mock_repo,
				mockUnleashClient,
			)
		})
	}
}

func TestVirtualClassRoomLogService_LogWhenUpdateRoomState(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	mock_repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID string
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:     "success case",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-id-1",
					repo.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
			hasError: false,
		},
		{
			name:     "failed case",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					"lesson-id-1",
					repo.TotalTimesUpdatingRoomState,
				).
					Return(errors.New("error")).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)
			v := &VirtualClassRoomLogService{
				WrapperConnection: wrapperConnection,
				Repo:              mock_repo,
			}
			err := v.LogWhenUpdateRoomState(ctx, tc.lessonID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				mock_repo,
				mockUnleashClient,
			)
		})
	}
}

func TestVirtualClassRoomLogService_LogWhenEndRoom(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	mock_repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID string
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:     "success case",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On(
					"CompleteLogByLessonID",
					ctx,
					db,
					"lesson-id-1",
				).
					Return(nil).
					Once()
			},
			hasError: false,
		},
		{
			name:     "failed case",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On(
					"CompleteLogByLessonID",
					ctx,
					db,
					"lesson-id-1",
				).
					Return(errors.New("error")).
					Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)
			v := &VirtualClassRoomLogService{
				WrapperConnection: wrapperConnection,
				Repo:              mock_repo,
			}
			err := v.LogWhenEndRoom(ctx, tc.lessonID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				mock_repo,
				mockUnleashClient,
			)
		})
	}
}

func TestVirtualClassRoomLogService_GetCompletedLogByLesson(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	mock_repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID string
		setup    func(ctx context.Context)
		hasLog   bool
		hasError bool
	}{
		{
			name:     "get completed log successfully",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On("GetLatestByLessonID", ctx, db, "lesson-id-1").
					Return(&repo.VirtualClassRoomLogDTO{
						LogID:       database.Text("log-id-1"),
						LessonID:    database.Text("lesson-id-1"),
						IsCompleted: database.Bool(true),
					}, nil).Once()
			},
			hasLog:   true,
			hasError: false,
		},
		{
			name:     "get completed log unsuccessfully",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On(
					"GetLatestByLessonID",
					ctx,
					db,
					"lesson-id-1",
				).
					Return(nil, errors.New("error")).
					Once()
			},
			hasError: true,
		},
		{
			name:     "get completed log successfully",
			lessonID: "lesson-id-1",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				mock_repo.On("GetLatestByLessonID", ctx, db, "lesson-id-1").
					Return(&repo.VirtualClassRoomLogDTO{
						LogID:       database.Text("log-id-1"),
						LessonID:    database.Text("lesson-id-1"),
						IsCompleted: database.Bool(false),
					}, nil).Once()
			},
			hasLog:   false,
			hasError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)
			v := &VirtualClassRoomLogService{
				WrapperConnection: wrapperConnection,
				Repo:              mock_repo,
			}
			actual, err := v.GetCompletedLogByLesson(ctx, tc.lessonID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.hasLog {
					assert.NotNil(t, actual)
				}
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				mock_repo,
				mockUnleashClient,
			)
		})
	}
}
