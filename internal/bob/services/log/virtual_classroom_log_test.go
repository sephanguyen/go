package log

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVirtualClassRoomLog_LogWhenAttendeeJoin(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name          string
		lessonID      pgtype.Text
		attendeeID    pgtype.Text
		setup         func(ctx context.Context)
		createdNewLog bool
		hasError      bool
	}{
		{
			name:       "new log successfully when these are no any current log",
			lessonID:   database.Text("lesson-id-1"),
			attendeeID: database.Text("user-id-1"),
			setup: func(ctx context.Context) {
				repo.On("GetLatestByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, pgx.ErrNoRows).Once()
				repo.On("Create", ctx, db, mock.Anything).Run(func(args mock.Arguments) {
					log := args[2].(*entities.VirtualClassRoomLog)
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
			lessonID:   database.Text("lesson-id-1"),
			attendeeID: database.Text("user-id-1"),
			setup: func(ctx context.Context) {
				repo.On("GetLatestByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(nil, errors.New("error")).Once()
			},
			hasError: true,
		},
		{
			name:       "new log successfully when current log completed",
			lessonID:   database.Text("lesson-id-1"),
			attendeeID: database.Text("user-id-1"),
			setup: func(ctx context.Context) {
				repo.On("GetLatestByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.VirtualClassRoomLog{
						LogID:       database.Text("log-id-1"),
						LessonID:    database.Text("lesson-id-1"),
						IsCompleted: database.Bool(true),
					}, nil).Once()
				repo.On("Create", ctx, db, mock.Anything).Run(func(args mock.Arguments) {
					log := args[2].(*entities.VirtualClassRoomLog)
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
			lessonID:   database.Text("lesson-id-1"),
			attendeeID: database.Text("user-id-1"),
			setup: func(ctx context.Context) {
				repo.On("GetLatestByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.VirtualClassRoomLog{
						LogID:       database.Text("log-id-1"),
						LessonID:    database.Text("lesson-id-1"),
						IsCompleted: database.Bool(false),
					}, nil).Once()
				repo.On("AddAttendeeIDByLessonID", ctx, db, database.Text("lesson-id-1"), database.Text("user-id-1")).
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
				DB:   db,
				Repo: repo,
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
				repo,
			)
		})
	}
}

func TestVirtualClassRoomLogService_LogWhenGetRoomState(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID pgtype.Text
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:     "success case",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-id-1"),
					entities.TotalTimesGettingRoomState,
				).
					Return(nil).
					Once()
			},
			hasError: false,
		},
		{
			name:     "failed case",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-id-1"),
					entities.TotalTimesGettingRoomState,
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
				DB:   db,
				Repo: repo,
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
				repo,
			)
		})
	}
}

func TestVirtualClassRoomLogService_LogWhenUpdateRoomState(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID pgtype.Text
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:     "success case",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-id-1"),
					entities.TotalTimesUpdatingRoomState,
				).
					Return(nil).
					Once()
			},
			hasError: false,
		},
		{
			name:     "failed case",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On(
					"IncreaseTotalTimesByLessonID",
					ctx,
					db,
					database.Text("lesson-id-1"),
					entities.TotalTimesUpdatingRoomState,
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
				DB:   db,
				Repo: repo,
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
				repo,
			)
		})
	}
}

func TestVirtualClassRoomLogService_LogWhenEndRoom(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID pgtype.Text
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name:     "success case",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On(
					"CompleteLogByLessonID",
					ctx,
					db,
					database.Text("lesson-id-1"),
				).
					Return(nil).
					Once()
			},
			hasError: false,
		},
		{
			name:     "failed case",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On(
					"CompleteLogByLessonID",
					ctx,
					db,
					database.Text("lesson-id-1"),
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
				DB:   db,
				Repo: repo,
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
				repo,
			)
		})
	}
}

func TestVirtualClassRoomLogService_GetCompletedLogByLesson(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}

	repo := new(mock_repositories.MockVirtualClassroomLogRepo)

	tcs := []struct {
		name     string
		lessonID pgtype.Text
		setup    func(ctx context.Context)
		hasLog   bool
		hasError bool
	}{
		{
			name:     "get completed log successfully",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On("GetLatestByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.VirtualClassRoomLog{
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
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On(
					"GetLatestByLessonID",
					ctx,
					db,
					database.Text("lesson-id-1"),
				).
					Return(nil, errors.New("error")).
					Once()
			},
			hasError: true,
		},
		{
			name:     "get completed log successfully",
			lessonID: database.Text("lesson-id-1"),
			setup: func(ctx context.Context) {
				repo.On("GetLatestByLessonID", ctx, db, database.Text("lesson-id-1")).
					Return(&entities.VirtualClassRoomLog{
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
				DB:   db,
				Repo: repo,
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
				repo,
			)
		})
	}
}
