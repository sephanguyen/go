package timesheet

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type ActionLogServiceImpl struct {
	DB database.Ext

	ActionLogRepo interface {
		Create(ctx context.Context, db database.QueryExecer, req *entity.TimesheetActionLog) error
	}
}

func (s *ActionLogServiceImpl) Create(ctx context.Context, req *tpb.TimesheetActionLogRequest) error {
	dto := dto.NewTimesheetActionLogDTOFromNATSRPCRequest(req)

	err := dto.ValidateCreateInfo()
	if err != nil {
		return err
	}

	actionLogEntity := dto.ToEntity()

	return s.ActionLogRepo.Create(ctx, s.DB, actionLogEntity)
}
