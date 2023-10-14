package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	nats_service_utils "github.com/manabie-com/backend/internal/timesheet/service/nats"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ServiceImpl struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	GetTimesheetService interface {
		GetTimesheet(ctx context.Context, timesheetQueryArgs *dto.TimesheetQueryArgs, timesheetQueryOptions *dto.TimesheetGetOptions) ([]*dto.Timesheet, error)
	}

	TimesheetRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, timesheets []*entity.Timesheet) ([]*entity.Timesheet, error)
		UpdateTimeSheet(ctx context.Context, db database.QueryExecer, timesheet *entity.Timesheet) (*entity.Timesheet, error)
		FindTimesheetByTimesheetID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Timesheet, error)
		FindTimesheetByTimesheetArgs(ctx context.Context, db database.QueryExecer, timesheetArgs *dto.TimesheetQueryArgs) ([]*entity.Timesheet, error)
		CountTimesheets(ctx context.Context, db database.QueryExecer, req *dto.TimesheetCountReq) (*dto.TimesheetCountOut, error)
		CountTimesheetsV2(ctx context.Context, db database.QueryExecer, req *dto.TimesheetCountV2Req) (*dto.TimesheetCountV2Out, error)
		GetTimesheetCountByStatusAndLocationIds(ctx context.Context, db database.QueryExecer, req *dto.TimesheetCountByStatusAndLocationIdsReq) (*dto.TimesheetCountByStatusAndLocationIdsResp, error)
	}

	OtherWorkingHoursRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, listOWHs []*entity.OtherWorkingHours) error
		FindListOtherWorkingHoursByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs pgtype.TextArray) ([]*entity.OtherWorkingHours, error)
	}

	TransportationExpenseRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, listTransportExpenses []*entity.TransportationExpense) error
		FindListTransportExpensesByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs pgtype.TextArray) ([]*entity.TransportationExpense, error)
	}

	TimesheetLessonHoursRepo interface {
		FindTimesheetLessonHoursByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) ([]*entity.TimesheetLessonHours, error)
		MapExistingLessonHoursByTimesheetIds(ctx context.Context, db database.QueryExecer, ids []string) (map[string]struct{}, error)
	}
}

func (s *ServiceImpl) CreateTimesheet(ctx context.Context, timesheet *dto.Timesheet) (string, error) {
	if err := checkPermissionToModifyTimesheet(ctx, timesheet.StaffID); err != nil {
		return "", status.Error(codes.PermissionDenied, err.Error())
	}

	timesheets, err := s.GetTimesheetService.GetTimesheet(ctx,
		&dto.TimesheetQueryArgs{
			StaffIDs:      []string{timesheet.StaffID},
			LocationID:    timesheet.LocationID,
			TimesheetDate: timesheet.TimesheetDate,
		},
		&dto.TimesheetGetOptions{
			IsGetListOtherWorkingHours:     true,
			IsGetListTimesheetLessonHours:  true,
			IsGetListTransportationExpense: true,
		},
	)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	var timesheetResult *dto.Timesheet
	switch len(timesheets) {
	case 0: // Timesheet not exist
		timesheetResult = s.buildCreateTimesheet(timesheet)
	case 1: // Timesheet already exists
		// Check Other Working Hours
		if len(timesheets[0].ListOtherWorkingHours) > 0 || len(timesheets[0].ListTransportationExpenses) > 0 {
			return "", status.Error(codes.AlreadyExists, constant.ErrorMessageDuplicateTimesheet)
		}

		// Check lesson hours contain valid lesson hours
		for _, elm := range timesheets[0].ListTimesheetLessonHours {
			if elm.FlagOn {
				return "", status.Error(codes.AlreadyExists, constant.ErrorMessageDuplicateTimesheet)
			}
		}

		timesheetResult, err = s.buildUpdateTimesheet(ctx, timesheets[0], timesheet)
		if err != nil {
			return "", status.Error(codes.Internal, err.Error())
		}

	default:
		return "", status.Error(codes.Internal, constant.ErrorMessageTooManyTimesheets)
	}

	var timesheetID string
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Create Timesheet Info
		timesheetE, err := s.TimesheetRepo.UpsertMultiple(ctx, tx, []*entity.Timesheet{timesheetResult.ToEntity()})
		if err != nil {
			return err
		}
		timesheetID = timesheetE[0].TimesheetID.String

		// create Other Working Hours Info
		if len(timesheetResult.ListOtherWorkingHours) > 0 {
			err = s.OtherWorkingHoursRepo.UpsertMultiple(ctx, tx, timesheetResult.ListOtherWorkingHours.ToEntities())
			if err != nil {
				return err
			}
		}

		// create Transportation Expenses Info
		if len(timesheetResult.ListTransportationExpenses) > 0 {
			err = s.TransportationExpenseRepo.UpsertMultiple(ctx, tx, timesheetResult.ListTransportationExpenses.ToEntities())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	return timesheetID, nil
}

func (s *ServiceImpl) UpdateTimesheet(ctx context.Context, req *dto.Timesheet) error {
	curTimesheet, err := s.TimesheetRepo.FindTimesheetByTimesheetID(ctx, s.DB, database.Text(req.ID))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", err.Error()))
	}

	if err := checkPermissionToModifyTimesheetWithTimesheetState(ctx, curTimesheet.StaffID.String, curTimesheet.TimesheetStatus.String); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	mapTSLessonHours, err := s.TimesheetLessonHoursRepo.MapExistingLessonHoursByTimesheetIds(ctx, s.DB, []string{req.ID})
	if err != nil {
		return status.Error(codes.Internal, fmt.Errorf("get list timesheet lesson hours by timesheet ids error: %v", err).Error())
	}

	// check timesheet empty
	if len(mapTSLessonHours) == 0 && len(req.ListOtherWorkingHours) == 0 && len(req.ListTransportationExpenses) == 0 {
		return status.Error(codes.Internal, "timesheet empty, cannot update anything")
	}

	updatedTimesheet, err := s.buildUpdateTimesheet(ctx, dto.NewTimesheetFromEntity(curTimesheet), req)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := s.TimesheetRepo.UpdateTimeSheet(ctx, tx, updatedTimesheet.ToEntity()); err != nil {
			return err
		}

		if len(updatedTimesheet.ListOtherWorkingHours) > 0 {
			err = s.OtherWorkingHoursRepo.UpsertMultiple(ctx, tx, updatedTimesheet.ListOtherWorkingHours.ToEntities())
			if err != nil {
				return err
			}
		}

		if len(updatedTimesheet.ListTransportationExpenses) > 0 {
			err = s.TransportationExpenseRepo.UpsertMultiple(ctx, tx, updatedTimesheet.ListTransportationExpenses.ToEntities())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// send manual edit timesheet event to NATS
	msg := &pb.TimesheetActionLogRequest{
		Action:      pb.TimesheetAction_EDITED,
		ExecutedBy:  interceptors.UserIDFromContext(ctx),
		TimesheetId: req.ID,
		IsSystem:    false,
		ExecutedAt:  timestamppb.New(time.Now()),
	}
	err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *ServiceImpl) CountTimesheets(ctx context.Context, req *dto.TimesheetCountReq) (*dto.TimesheetCountOut, error) {
	return s.TimesheetRepo.CountTimesheets(ctx, s.DB, req)
}

func (s *ServiceImpl) CountTimesheetsV2(ctx context.Context, req *dto.TimesheetCountV2Req) (*dto.TimesheetCountV2Out, error) {
	return s.TimesheetRepo.CountTimesheetsV2(ctx, s.DB, req)
}

func (s *ServiceImpl) CountSubmittedTimesheets(ctx context.Context, req *dto.CountSubmittedTimesheetsReq) (*dto.CountSubmittedTimesheetsResp, error) {
	res, err := s.TimesheetRepo.GetTimesheetCountByStatusAndLocationIds(ctx, s.DB, &dto.TimesheetCountByStatusAndLocationIdsReq{
		Status:      pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String(),
		FromDate:    dto.GetMinTimesheetDateValid(),
		ToDate:      dto.GetTimesheetEndOfMonthDate(),
		LocationIds: req.LocationIds,
	})

	// return error, skip if its a "no row error"
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	// return 0 if error is a "no row error"
	if err == pgx.ErrNoRows {
		return &dto.CountSubmittedTimesheetsResp{
			Count: 0,
		}, nil
	}

	return &dto.CountSubmittedTimesheetsResp{
		Count: res.Count,
	}, nil
}

func checkPermissionToModifyTimesheetWithTimesheetState(ctx context.Context, timesheetStaffID, timesheetStatus string) error {
	userRoles := interceptors.UserRolesFromContext(ctx)
	userID := interceptors.UserIDFromContext(ctx)
	switch timesheetStatus {
	case pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String():
		// both requester and approver/confirmer
		if timesheetStaffID == userID {
			return nil
		}

		for _, role := range userRoles {
			if _, found := constant.RolesWriteOtherMemberTimesheet[role]; found {
				return nil
			}
		}
	case pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String():
		// check if approver or confirmer
		for _, role := range userRoles {
			if _, found := constant.RolesWriteOtherMemberTimesheet[role]; found {
				return nil
			}
		}
	default:
		return fmt.Errorf("unauthorized to modify timesheet, timesheetStaffID: %s", timesheetStaffID)
	}

	return fmt.Errorf("unauthorized to modify timesheet, timesheetStaffID: %s", timesheetStaffID)
}

func checkPermissionToModifyTimesheet(ctx context.Context, timesheetStaffID string) error {
	userRoles := interceptors.UserRolesFromContext(ctx)
	userID := interceptors.UserIDFromContext(ctx)

	if timesheetStaffID == userID {
		return nil
	}

	for _, role := range userRoles {
		if _, found := constant.RolesWriteOtherMemberTimesheet[role]; found {
			return nil
		}
	}

	return fmt.Errorf("unauthorized to modify timesheet, timesheetStaffID: %s", timesheetStaffID)
}

func (s *ServiceImpl) buildCreateTimesheet(timesheet *dto.Timesheet) *dto.Timesheet {
	// create new timesheet info
	timesheet.MakeNewID()
	timesheet.TimesheetStatus = pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()

	// create new timesheet other working hours
	if len(timesheet.ListOtherWorkingHours) > 0 {
		timesheet.ListOtherWorkingHours = buildListOtherWorkingHours(timesheet.ID, timesheet.ListOtherWorkingHours, true /*isCreate*/)
	}

	// create new transport expense data for timesheet
	if len(timesheet.ListTransportationExpenses) > 0 {
		timesheet.ListTransportationExpenses = buildListTransportExpenses(timesheet.ID, timesheet.ListTransportationExpenses, true /*isCreate*/)
	}

	return timesheet
}

func (s *ServiceImpl) buildUpdateTimesheet(ctx context.Context, cur *dto.Timesheet, new *dto.Timesheet) (*dto.Timesheet, error) {
	// build timesheet info
	timesheet := buildUpdateTimesheetInfo(cur, new)

	// build timesheet other working hours
	listCurrentOWHs, err := s.OtherWorkingHoursRepo.FindListOtherWorkingHoursByTimesheetIDs(ctx, s.DB, database.TextArray([]string{timesheet.ID}))
	if err != nil {
		return nil, fmt.Errorf("get list other working hours error: %s", err.Error())
	}

	listCurrentTransportExpenses, err := s.TransportationExpenseRepo.FindListTransportExpensesByTimesheetIDs(ctx, s.DB, database.TextArray([]string{timesheet.ID}))
	if err != nil {
		return nil, fmt.Errorf("get list transport expenses error: %s", err.Error())
	}

	listUpdateOWHsReq := calculateListOtherWorkingHoursUpdate(listCurrentOWHs, new.ListOtherWorkingHours)
	listUpdateTransportExpensesReq := calculateListTransportExpensesUpdate(listCurrentTransportExpenses, new.ListTransportationExpenses)

	timesheet.ListOtherWorkingHours = buildListOtherWorkingHours(timesheet.ID, listUpdateOWHsReq, false /*isCreate*/)
	timesheet.ListTransportationExpenses = buildListTransportExpenses(timesheet.ID, listUpdateTransportExpensesReq, false /*isCreate*/)

	return timesheet, nil
}

func buildUpdateTimesheetInfo(cur *dto.Timesheet, new *dto.Timesheet) *dto.Timesheet {
	cur.Remark = new.Remark
	return cur
}

func buildListOtherWorkingHours(timesheetID string, listOWHsRequest []*dto.OtherWorkingHours, isCreate bool) []*dto.OtherWorkingHours {
	for i := range listOWHsRequest {
		listOWHsRequest[i].NormalizedData()
		listOWHsRequest[i].TimesheetID = timesheetID

		// create new Other Working Hours ID
		if isCreate || listOWHsRequest[i].ID == "" {
			listOWHsRequest[i].ID = idutil.ULIDNow()
		}

		// calculate total hours - minute (int)
		listOWHsRequest[i].TotalHour = int16(listOWHsRequest[i].EndTime.Sub(listOWHsRequest[i].StartTime).Minutes())
	}
	return listOWHsRequest
}

func calculateListOtherWorkingHoursUpdate(current []*entity.OtherWorkingHours, reqNew []*dto.OtherWorkingHours) []*dto.OtherWorkingHours {
	currentDto := dto.NewListOtherWorkingHoursFromEntities(current)

	mapNewOWHsIDs := make(map[string]struct{})
	for _, elm := range reqNew {
		mapNewOWHsIDs[elm.ID] = struct{}{}
	}

	for _, elm := range currentDto {
		if _, found := mapNewOWHsIDs[elm.ID]; !found {
			elm.IsDeleted = true
			reqNew = append(reqNew, elm)
		}
	}

	return reqNew
}

func buildListTransportExpenses(timesheetID string, listTransportExpensesRequest []*dto.TransportationExpenses, isCreate bool) []*dto.TransportationExpenses {
	for i := range listTransportExpensesRequest {
		listTransportExpensesRequest[i].TimesheetID = timesheetID

		// create new Other Working Hours ID
		if isCreate || listTransportExpensesRequest[i].TransportExpenseID == "" {
			listTransportExpensesRequest[i].TransportExpenseID = idutil.ULIDNow()
		}
	}
	return listTransportExpensesRequest
}

func calculateListTransportExpensesUpdate(current []*entity.TransportationExpense, reqNew []*dto.TransportationExpenses) []*dto.TransportationExpenses {
	currentDto := dto.NewListTransportExpensesFromEntities(current)

	mapNewTransportExpensesIDs := make(map[string]struct{})
	for _, elm := range reqNew {
		mapNewTransportExpensesIDs[elm.TransportExpenseID] = struct{}{}
	}

	for _, elm := range currentDto {
		if _, found := mapNewTransportExpensesIDs[elm.TransportExpenseID]; !found {
			elm.IsDeleted = true
			reqNew = append(reqNew, elm)
		}
	}

	return reqNew
}
