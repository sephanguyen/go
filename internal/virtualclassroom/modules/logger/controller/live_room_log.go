package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type LiveRoomLogService struct {
	DB database.Ext

	LiveRoomLogRepo infrastructure.LiveRoomLogRepo
}

func (l *LiveRoomLogService) LogWhenAttendeeJoin(ctx context.Context, channelID, attendeeID string) (createdNewLog bool, err error) {
	log, err := l.LiveRoomLogRepo.GetLatestByChannelID(ctx, l.DB, channelID)
	if errors.Is(err, pgx.ErrNoRows) {
		createdNewLog = true
	} else if err != nil {
		return false, fmt.Errorf("error in LiveRoomLogRepo.GetLatestByChannelID, channel %s: %w", channelID, err)
	}

	// create new log
	if createdNewLog || log.IsCompleted.Bool {
		createdNewLog = true
		dto := &repo.LiveRoomLog{}
		database.AllNullEntity(dto)
		if err = multierr.Combine(
			dto.LiveRoomLogID.Set(idutil.ULIDNow()),
			dto.ChannelID.Set(channelID),
			dto.IsCompleted.Set(false),
			dto.AttendeeIDs.Set([]string{attendeeID}),
		); err != nil {
			return false, fmt.Errorf("could not set value for entity LiveRoomLog: %w", err)
		}

		if err = l.LiveRoomLogRepo.Create(ctx, l.DB, dto); err != nil {
			return false, fmt.Errorf("error in LiveRoomLogRepo.Create, channel %s: %w", channelID, err)
		}
	} else {
		if err = l.LiveRoomLogRepo.AddAttendeeIDByChannelID(ctx, l.DB, channelID, attendeeID); err != nil {
			return false, fmt.Errorf("error in LiveRoomLogRepo.AddAttendeeIDByChannelID, channel %s, user %s: %w", channelID, attendeeID, err)
		}
	}

	return createdNewLog, nil
}

func (l *LiveRoomLogService) LogWhenUpdateRoomState(ctx context.Context, channelID string) error {
	if err := l.LiveRoomLogRepo.IncreaseTotalTimesByChannelID(ctx, l.DB, channelID, repo.TotalTimesUpdatingRoomState); err != nil {
		return fmt.Errorf("error in LiveRoomLogRepo.IncreaseTotalTimesByChannelID, channel %s: %w", channelID, err)
	}

	return nil
}

func (l *LiveRoomLogService) LogWhenGetRoomState(ctx context.Context, channelID string) error {
	if err := l.LiveRoomLogRepo.IncreaseTotalTimesByChannelID(ctx, l.DB, channelID, repo.TotalTimesGettingRoomState); err != nil {
		return fmt.Errorf("error in LiveRoomLogRepo.IncreaseTotalTimesByChannelID, channel %s: %w", channelID, err)
	}

	return nil
}

func (l *LiveRoomLogService) LogWhenEndRoom(ctx context.Context, channelID string) error {
	if err := l.LiveRoomLogRepo.CompleteLogByChannelID(ctx, l.DB, channelID); err != nil {
		return fmt.Errorf("error in LiveRoomLogRepo.CompleteLogByChannelID, channel %s: %w", channelID, err)
	}

	return nil
}

func (l *LiveRoomLogService) GetCompletedLogByChannel(ctx context.Context, channelID string) (*repo.LiveRoomLog, error) {
	log, err := l.LiveRoomLogRepo.GetLatestByChannelID(ctx, l.DB, channelID)
	if err != nil {
		return nil, fmt.Errorf("error in LiveRoomLogRepo.GetLatestByChannelID, channel %s: %w", channelID, err)
	}

	if !log.IsCompleted.Bool {
		return nil, nil
	}

	return log, nil
}
