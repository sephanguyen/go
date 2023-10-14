package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	nats_service_utils "github.com/manabie-com/backend/internal/timesheet/service/nats"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type StaffTransportationExpenseServiceImpl struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	StaffTransportationExpenseRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, listStaffTransportExpenses []*entity.StaffTransportationExpense) error
		FindListTransportExpensesByStaffIDs(ctx context.Context, db database.QueryExecer, StaffIDs pgtype.TextArray) ([]*entity.StaffTransportationExpense, error)
	}

	TimesheetRepo interface {
		GetStaffFutureTimesheetIDsWithLocations(ctx context.Context, db database.QueryExecer, staffID string, date time.Time, locationIDs []string) ([]dto.TimesheetLocationDto, error)
	}

	TransportationExpenseRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, listTransportExpenses []*entity.TransportationExpense) error
		SoftDeleteMultipleByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) error
	}
}

func (s *StaffTransportationExpenseServiceImpl) UpsertConfig(ctx context.Context, staffID string, listStaffTransportationExpenseRequest *dto.ListStaffTransportationExpenses) error {
	listCurrentStaffTransportExpenses, err := s.StaffTransportationExpenseRepo.FindListTransportExpensesByStaffIDs(ctx, s.DB, database.TextArray([]string{staffID}))
	if err != nil {
		return fmt.Errorf("get list staff transport expenses error: %s", err.Error())
	}
	listUpdateTransportExpensesReq, tsLocationIDs := calculateStaffListTransportExpensesUpdate(listCurrentStaffTransportExpenses, listStaffTransportationExpenseRequest)

	listStaffTransportationExpenseResult := buildListStaffTransportationExpense(listUpdateTransportExpensesReq)

	// get future timesheet can change TEs
	// with update TE records:
	//   - same staff TE location -> update all future ts
	//   - change staff TE location -> update future ts of new location
	//   - remove staff TE, not do anything
	timesheetIDs, err := s.TimesheetRepo.GetStaffFutureTimesheetIDsWithLocations(ctx, s.DB, staffID, time.Now().In(timeutil.Timezone(pbc.COUNTRY_JP)), tsLocationIDs)
	if err != nil {
		return fmt.Errorf("get staff timesheet after date error: %s", err.Error())
	}

	listDeleteTS := []string{}
	for _, tsLocation := range timesheetIDs {
		listDeleteTS = append(listDeleteTS, tsLocation.TimesheetID)
	}

	TEs := buildNewTEsForStaff(timesheetIDs, listStaffTransportationExpenseResult)

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := s.StaffTransportationExpenseRepo.UpsertMultiple(ctx, tx, listStaffTransportationExpenseResult.ToEntities())
		if err != nil {
			return err
		}

		err = s.TransportationExpenseRepo.SoftDeleteMultipleByTimesheetIDs(ctx, tx, listDeleteTS)
		if err != nil {
			return err
		}

		err = s.TransportationExpenseRepo.UpsertMultiple(ctx, tx, TEs.ToEntities())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	timeExecuted := time.Now()
	// keep track of timesheet ids that already sent action log event
	sentActionLogTimesheetIDs := make(map[string]bool)
	for _, transportationExpense := range TEs.ToEntities() {
		timesheetID := transportationExpense.TimesheetID.String
		// send action log event
		if _, ok := sentActionLogTimesheetIDs[timesheetID]; !ok {
			msg := &pb.TimesheetActionLogRequest{
				Action:      pb.TimesheetAction_EDITED,
				ExecutedBy:  interceptors.UserIDFromContext(ctx),
				TimesheetId: timesheetID,
				IsSystem:    true,
				ExecutedAt:  timestamppb.New(timeExecuted),
			}
			err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}
			sentActionLogTimesheetIDs[timesheetID] = true
		}
	}

	return nil
}

func buildNewTEsForStaff(tsAndLocationIDs []dto.TimesheetLocationDto, staffTEs *dto.ListStaffTransportationExpenses) dto.ListTransportationExpenses {
	TEs := make([]*dto.TransportationExpenses, 0)
	for _, tsAndLocationID := range tsAndLocationIDs {
		for _, staffTE := range *staffTEs {
			if !staffTE.IsDeleted && tsAndLocationID.LocationID == staffTE.LocationID {
				te := &dto.TransportationExpenses{
					TransportExpenseID: idutil.ULIDNow(),
					TimesheetID:        tsAndLocationID.TimesheetID,
					TransportationType: staffTE.TransportationType,
					TransportationFrom: staffTE.TransportationFrom,
					TransportationTo:   staffTE.TransportationTo,
					CostAmount:         staffTE.CostAmount,
					RoundTrip:          staffTE.RoundTrip,
					Remarks:            staffTE.Remarks,
					IsDeleted:          staffTE.IsDeleted, // alway is "false"
				}

				TEs = append(TEs, te)
			}
		}
	}

	return TEs
}

func buildListStaffTransportationExpense(listStaffTransportationExpense []*dto.StaffTransportationExpenses) *dto.ListStaffTransportationExpenses {
	listStaffTransportationExpenseResult := make(dto.ListStaffTransportationExpenses, 0, len(listStaffTransportationExpense))

	for i := range listStaffTransportationExpense {
		if listStaffTransportationExpense[i].ID == "" {
			listStaffTransportationExpense[i].ID = idutil.ULIDNow()
		}

		listStaffTransportationExpenseResult = append(listStaffTransportationExpenseResult, listStaffTransportationExpense[i])
	}

	return &listStaffTransportationExpenseResult
}

func calculateStaffListTransportExpensesUpdate(current []*entity.StaffTransportationExpense, reqNew *dto.ListStaffTransportationExpenses) ([]*dto.StaffTransportationExpenses, []string) {
	currentDto := dto.NewListStaffTransportExpensesFromEntities(current)

	mapTSLocationIDsNeedUpdate := make(map[string]struct{})
	mapNewTransportExpensesIDs := make(map[string]*dto.StaffTransportationExpenses)
	for _, elm := range *reqNew {
		mapNewTransportExpensesIDs[elm.ID] = elm

		if len(elm.ID) == 0 {
			mapTSLocationIDsNeedUpdate[elm.LocationID] = struct{}{}
		}
	}

	for _, curTE := range currentDto {
		if newTE, found := mapNewTransportExpensesIDs[curTE.ID]; !found {
			// TE deleted
			curTE.IsDeleted = true
			*reqNew = append(*reqNew, curTE)
		} else {
			if curTE.LocationID != newTE.LocationID { // update location need remove all timesheet TEs at new location and add new
				mapTSLocationIDsNeedUpdate[newTE.LocationID] = struct{}{}
			} else if !curTE.IsEqual(newTE) { // update other fields -> need removed all old timesheet TEs and add new
				mapTSLocationIDsNeedUpdate[curTE.LocationID] = struct{}{}
			}
			// not have any update, not do anything
		}
	}

	// save map to slice
	listTSLocationIDsNeedUpdate := make([]string, len(mapTSLocationIDsNeedUpdate))
	for locationID := range mapTSLocationIDsNeedUpdate {
		listTSLocationIDsNeedUpdate = append(listTSLocationIDsNeedUpdate, locationID)
	}

	return *reqNew, listTSLocationIDsNeedUpdate
}
