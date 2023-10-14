package services

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/zeus/entities"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	zpb "github.com/manabie-com/backend/pkg/manabuf/zeus/v1"
)

type CentralizeLogsService struct {
	DB database.Ext
	zpb.UnimplementedCentralizeLogsServiceServer
	ActivityLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, en *entities.ActivityLog) error
		CreateBulk(ctx context.Context, db database.QueryExecer, logs []*entities.ActivityLog) error
	}
}

func (s *CentralizeLogsService) CreateLogs(ctx context.Context, msg *npb.ActivityLogEvtCreated) error {
	activityLog, err := entities.ToActivityLog(msg.UserId, msg.ActionType, string(msg.Payload), msg.ResourcePath, msg.RequestAt.AsTime(), msg.Status, msg.FinishedAt.AsTime())
	if err != nil {
		return err
	}
	return s.ActivityLogRepo.Create(ctx, s.DB, activityLog)
}

func (s *CentralizeLogsService) BulkCreateLogs(ctx context.Context, logs []*entities.ActivityLog) error {
	return s.ActivityLogRepo.CreateBulk(ctx, s.DB, logs)
}
