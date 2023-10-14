package dto

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

type TimesheetActionLogReq struct {
	TimesheetID string
	UserID      string
	IsSystem    bool
	Action      string
	ExecutedAt  time.Time
}

func NewTimesheetActionLogDTOFromNATSRPCRequest(req *tpb.TimesheetActionLogRequest) *TimesheetActionLogReq {
	return &TimesheetActionLogReq{
		TimesheetID: req.GetTimesheetId(),
		UserID:      req.GetExecutedBy(),
		IsSystem:    req.GetIsSystem(),
		Action:      req.GetAction().String(),
		ExecutedAt:  req.GetExecutedAt().AsTime(),
	}
}

func (t *TimesheetActionLogReq) ValidateCreateInfo() error {
	switch {
	case t.TimesheetID == "":
		return fmt.Errorf("timesheet id must not be empty")
	case t.UserID == "" && !t.IsSystem:
		return fmt.Errorf("user id must not be empty")
	case t.Action == "":
		return fmt.Errorf("action must not be empty")
	case t.ExecutedAt.IsZero():
		return fmt.Errorf("executed at must not be empty")
	}
	return nil
}

func (t *TimesheetActionLogReq) ToEntity() *entity.TimesheetActionLog {
	return &entity.TimesheetActionLog{
		ID:          database.Text(idutil.ULIDNow()),
		TimesheetID: database.Text(t.TimesheetID),
		UserID:      database.Text(t.UserID),
		IsSystem:    database.Bool(t.IsSystem),
		Action:      database.Text(t.Action),
		ExecutedAt:  database.Timestamptz(t.ExecutedAt),
	}
}
