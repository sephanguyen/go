package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/calendar/application/queries/payloads"
	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type GetStaff struct {
	UserRepo infrastructure.UserPort
}

func (t *GetStaff) GetStaffsByLocation(ctx context.Context, db database.QueryExecer, req *payloads.GetStaffRequest) (*payloads.GetStaffResponse, error) {
	workingStatus := []string{
		string(constants.Available),
		string(constants.OnLeave),
	}

	staffs, err := t.UserRepo.GetStaffsByLocationAndWorkingStatus(ctx, db, req.LocationID, workingStatus, req.IsUsingUserBasicInfoTable)
	if err != nil {
		return nil, err
	}

	return &payloads.GetStaffResponse{
		User: staffs,
	}, nil
}

func (t *GetStaff) GetStaffsByLocationIDsAndNameOrEmail(ctx context.Context, db database.QueryExecer, req *payloads.GetStaffByLocationIDsAndNameOrEmailRequest) (*payloads.GetStaffByLocationIDsAndNameOrEmailResponse, error) {
	staffs, err := t.UserRepo.GetStaffsByLocationIDsAndNameOrEmail(ctx, db, req.LocationIDs, req.FilteredTeacherIDs, req.Keyword, req.Limit)
	if err != nil {
		return nil, err
	}

	return &payloads.GetStaffByLocationIDsAndNameOrEmailResponse{
		User: staffs,
	}, nil
}
