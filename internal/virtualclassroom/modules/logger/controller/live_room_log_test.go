package controller

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/liveroom/repositories"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLiveRoomLogService_LogWhenAttendeeJoin(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	mockLiveRoomLogRepo := new(mock_repositories.MockLiveRoomLogRepo)

	logID := "log-id1"
	channelID := "channel-id1"
	userID := "user-id1"

	tcs := []struct {
		name          string
		channelID     string
		attendeeID    string
		setup         func(ctx context.Context)
		createdNewLog bool
		hasError      bool
	}{
		{
			name:       "new log successfully when there are no any current log",
			channelID:  channelID,
			attendeeID: userID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(nil, pgx.ErrNoRows).Once()

				mockLiveRoomLogRepo.On("Create", ctx, db, mock.Anything).Run(func(args mock.Arguments) {
					log := args[2].(*repo.LiveRoomLog)
					assert.NotEmpty(t, log.LiveRoomLogID.String)
					assert.Equal(t, channelID, log.ChannelID.String)
					assert.False(t, log.IsCompleted.Bool)
					assert.ElementsMatch(t, database.FromTextArray(log.AttendeeIDs), []string{userID})
				}).Return(nil).Once()
			},
			createdNewLog: true,
			hasError:      false,
		},
		{
			name:       "get latest log fail",
			channelID:  channelID,
			attendeeID: userID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(nil, errors.New("error")).Once()
			},
			hasError: true,
		},
		{
			name:       "new log successfully when current log completed",
			channelID:  channelID,
			attendeeID: userID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(true),
					}, nil).Once()

				mockLiveRoomLogRepo.On("Create", ctx, db, mock.Anything).Run(func(args mock.Arguments) {
					log := args[2].(*repo.LiveRoomLog)
					assert.NotEmpty(t, log.LiveRoomLogID.String)
					assert.Equal(t, channelID, log.ChannelID.String)
					assert.False(t, log.IsCompleted.Bool)
					assert.ElementsMatch(t, database.FromTextArray(log.AttendeeIDs), []string{userID})
				}).Return(nil).Once()
			},
			createdNewLog: true,
			hasError:      false,
		},
		{
			name:       "update log successfully",
			channelID:  channelID,
			attendeeID: userID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(false),
					}, nil).Once()

				mockLiveRoomLogRepo.On("AddAttendeeIDByChannelID", ctx, db, channelID, userID).
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

			service := &LiveRoomLogService{
				DB:              db,
				LiveRoomLogRepo: mockLiveRoomLogRepo,
			}

			createdNewLogActual, err := service.LogWhenAttendeeJoin(ctx, tc.channelID, tc.attendeeID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tc.createdNewLog, createdNewLogActual)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockLiveRoomLogRepo)
		})
	}
}

func TestLiveRoomLogService_LogWhenUpdateRoomState(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	mockLiveRoomLogRepo := new(mock_repositories.MockLiveRoomLogRepo)

	channelID := "channel-id1"

	tcs := []struct {
		name      string
		channelID string
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "success case",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID,
					repo.TotalTimesUpdatingRoomState).
					Return(nil).Once()
			},
			hasError: false,
		},
		{
			name:      "failed case",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID,
					repo.TotalTimesUpdatingRoomState).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)

			service := &LiveRoomLogService{
				DB:              db,
				LiveRoomLogRepo: mockLiveRoomLogRepo,
			}

			err := service.LogWhenUpdateRoomState(ctx, tc.channelID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockLiveRoomLogRepo)
		})
	}
}

func TestLiveRoomLogService_LogWhenGetRoomState(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	mockLiveRoomLogRepo := new(mock_repositories.MockLiveRoomLogRepo)

	channelID := "channel-id1"

	tcs := []struct {
		name      string
		channelID string
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "success case",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID,
					repo.TotalTimesGettingRoomState).
					Return(nil).Once()
			},
			hasError: false,
		},
		{
			name:      "failed case",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("IncreaseTotalTimesByChannelID", ctx, db, channelID,
					repo.TotalTimesGettingRoomState).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)

			service := &LiveRoomLogService{
				DB:              db,
				LiveRoomLogRepo: mockLiveRoomLogRepo,
			}

			err := service.LogWhenGetRoomState(ctx, tc.channelID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockLiveRoomLogRepo)
		})
	}
}

func TestLiveRoomLogService_LogWhenEndRoom(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	mockLiveRoomLogRepo := new(mock_repositories.MockLiveRoomLogRepo)

	channelID := "channel-id1"

	tcs := []struct {
		name      string
		channelID string
		setup     func(ctx context.Context)
		hasError  bool
	}{
		{
			name:      "success case",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("CompleteLogByChannelID", ctx, db, channelID).
					Return(nil).Once()
			},
			hasError: false,
		},
		{
			name:      "failed case",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("CompleteLogByChannelID", ctx, db, channelID).
					Return(errors.New("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)

			service := &LiveRoomLogService{
				DB:              db,
				LiveRoomLogRepo: mockLiveRoomLogRepo,
			}

			err := service.LogWhenEndRoom(ctx, tc.channelID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, db, mockLiveRoomLogRepo)
		})
	}
}

func TestLiveRoomLogService_GetCompletedLogByChannel(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	mockLiveRoomLogRepo := new(mock_repositories.MockLiveRoomLogRepo)

	logID := "log-id1"
	channelID := "channel-id1"

	tcs := []struct {
		name      string
		channelID string
		setup     func(ctx context.Context)
		hasLog    bool
		hasError  bool
	}{
		{
			name:      "get completed log successfully",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(&repo.LiveRoomLog{
						LiveRoomLogID: database.Text(logID),
						ChannelID:     database.Text(channelID),
						IsCompleted:   database.Bool(true),
					}, nil).Once()
			},
			hasLog:   true,
			hasError: false,
		},
		{
			name:      "failed to get completed log",
			channelID: channelID,
			setup: func(ctx context.Context) {
				mockLiveRoomLogRepo.On("GetLatestByChannelID", ctx, db, channelID).
					Return(nil, errors.New("error")).Once()
			},
			hasLog:   false,
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.setup(ctx)

			service := &LiveRoomLogService{
				DB:              db,
				LiveRoomLogRepo: mockLiveRoomLogRepo,
			}

			actual, err := service.GetCompletedLogByChannel(ctx, tc.channelID)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.hasLog {
					assert.NotNil(t, actual)
				}
			}

			mock.AssertExpectationsForObjects(t, db, mockLiveRoomLogRepo)
		})
	}
}
