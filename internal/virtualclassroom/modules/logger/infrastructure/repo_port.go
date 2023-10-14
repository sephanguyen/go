package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
)

type VirtualClassroomLogRepo interface {
	Create(ctx context.Context, db database.QueryExecer, e *repo.VirtualClassRoomLogDTO) error
	AddAttendeeIDByLessonID(ctx context.Context, db database.QueryExecer, lessonID, attendeeID string) error
	IncreaseTotalTimesByLessonID(ctx context.Context, db database.QueryExecer, lessonID string, logType repo.TotalTimes) error
	CompleteLogByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) error
	GetLatestByLessonID(ctx context.Context, db database.QueryExecer, lessonID string) (*repo.VirtualClassRoomLogDTO, error)
}

type LiveRoomLogRepo interface {
	Create(ctx context.Context, db database.QueryExecer, dto *repo.LiveRoomLog) error
	AddAttendeeIDByChannelID(ctx context.Context, db database.QueryExecer, channelID, attendeeID string) error
	IncreaseTotalTimesByChannelID(ctx context.Context, db database.QueryExecer, channelID string, logType repo.TotalTimes) error
	CompleteLogByChannelID(ctx context.Context, db database.QueryExecer, channelID string) error
	GetLatestByChannelID(ctx context.Context, db database.QueryExecer, channelID string) (*repo.LiveRoomLog, error)
}
