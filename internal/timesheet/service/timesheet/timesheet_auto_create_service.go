package timesheet

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type AutoCreateTimesheetServiceImpl struct {
	DB database.Ext

	TimesheetRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, timesheets []*entity.Timesheet) ([]*entity.Timesheet, error)
		SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, id pgtype.TextArray) error
	}

	TimesheetLessonHoursRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, listTimesheetLessonHours []*entity.TimesheetLessonHours) ([]*entity.TimesheetLessonHours, error)
		SoftDelete(ctx context.Context, db database.QueryExecer, listTimesheetLessonHours []*entity.TimesheetLessonHours) error
		UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string, flagOn bool) error
	}

	TransportationExpenseRepo interface {
		UpsertMultiple(ctx context.Context, db database.QueryExecer, listTransportExpenses []*entity.TransportationExpense) error
		SoftDeleteMultipleByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) error
	}

	StaffTransportationExpenseRepo interface {
		FindListTransportExpensesByStaffIDsAndLocation(ctx context.Context, db database.QueryExecer, staffIDs []string, location string) (map[string][]entity.StaffTransportationExpense, error)
	}
}

func (s *AutoCreateTimesheetServiceImpl) CreateTimesheetMultiple(ctx context.Context, timesheets []*dto.Timesheet) ([]*dto.Timesheet, error) {
	newTimesheetEntities, newTimesheetLessonHoursEntities := buildNewTimesheets(timesheets)

	if len(newTimesheetEntities) == 0 && len(newTimesheetLessonHoursEntities) == 0 {
		return timesheets, nil
	}

	newTEs, err := s.buildTransportExpense(ctx, newTimesheetEntities)
	if err != nil {
		return nil, err
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if len(newTimesheetEntities) > 0 {
			if _, err := s.TimesheetRepo.UpsertMultiple(ctx, tx, newTimesheetEntities); err != nil {
				return err
			}
		}

		if len(newTimesheetLessonHoursEntities) > 0 {
			if _, err := s.TimesheetLessonHoursRepo.UpsertMultiple(ctx, tx, newTimesheetLessonHoursEntities); err != nil {
				return err
			}
		}

		if len(newTEs) > 0 {
			if err := s.TransportationExpenseRepo.UpsertMultiple(ctx, tx, newTEs); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return timesheets, err
	}
	return timesheets, nil
}

func (s *AutoCreateTimesheetServiceImpl) RemoveTimesheetLessonHoursMultiple(ctx context.Context, timesheets []*dto.Timesheet) ([]*dto.Timesheet, error) {
	removeTimesheetIDs, removeTimesheetLessonHoursEntities := buildRemoveTimesheets(timesheets)

	if len(removeTimesheetIDs) == 0 && len(removeTimesheetLessonHoursEntities) == 0 {
		return timesheets, nil
	}

	if err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if len(removeTimesheetLessonHoursEntities) > 0 {
			if err := s.TimesheetLessonHoursRepo.SoftDelete(ctx, s.DB, removeTimesheetLessonHoursEntities); err != nil {
				return err
			}
		}

		if len(removeTimesheetIDs) > 0 {
			if err := s.TimesheetRepo.SoftDeleteByIDs(ctx, s.DB, database.TextArray(removeTimesheetIDs)); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return timesheets, nil
}

func (s *AutoCreateTimesheetServiceImpl) CreateAndRemoveTimesheetMultiple(ctx context.Context, newTimesheets, removeTimesheets []*dto.Timesheet) error {
	newTimesheetEntities, newTimesheetLessonHoursEntities := buildNewTimesheets(newTimesheets)
	removeTimesheetIDs, removeTimesheetLessonHoursEntities := buildRemoveTimesheets(removeTimesheets)

	if len(newTimesheetEntities) == 0 &&
		len(newTimesheetLessonHoursEntities) == 0 &&
		len(removeTimesheetIDs) == 0 &&
		len(removeTimesheetLessonHoursEntities) == 0 {
		return nil
	}

	newTEs, err := s.buildTransportExpense(ctx, newTimesheetEntities)
	if err != nil {
		return err
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if len(newTimesheetEntities) > 0 {
			if _, err := s.TimesheetRepo.UpsertMultiple(ctx, tx, newTimesheetEntities); err != nil {
				return err
			}
		}

		if len(newTimesheetLessonHoursEntities) > 0 {
			if _, err := s.TimesheetLessonHoursRepo.UpsertMultiple(ctx, tx, newTimesheetLessonHoursEntities); err != nil {
				return err
			}
		}

		if len(newTEs) > 0 {
			if err := s.TransportationExpenseRepo.UpsertMultiple(ctx, tx, newTEs); err != nil {
				return err
			}
		}

		if len(removeTimesheetLessonHoursEntities) > 0 {
			if err := s.TimesheetLessonHoursRepo.SoftDelete(ctx, tx, removeTimesheetLessonHoursEntities); err != nil {
				return err
			}
		}

		if len(removeTimesheetIDs) > 0 {
			if err := s.TimesheetRepo.SoftDeleteByIDs(ctx, tx, database.TextArray(removeTimesheetIDs)); err != nil {
				return err
			}
			if err := s.TransportationExpenseRepo.SoftDeleteMultipleByTimesheetIDs(ctx, tx, removeTimesheetIDs); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *AutoCreateTimesheetServiceImpl) UpdateLessonAutoCreateFlagState(ctx context.Context, flagsMap map[bool][]string) error {
	return database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for flagOn, timesheetIDs := range flagsMap {
			err := s.TimesheetLessonHoursRepo.UpdateTimesheetLessonAutoCreateFlagByTimesheetIDs(ctx, tx, timesheetIDs, flagOn)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func buildNewTimesheets(timesheets []*dto.Timesheet) ([]*entity.Timesheet, []*entity.TimesheetLessonHours) {
	var (
		newTimesheetEntities            []*entity.Timesheet
		newTimesheetLessonHoursEntities []*entity.TimesheetLessonHours
	)

	for _, timesheet := range timesheets {
		timesheet.NormalizedData()
		if !timesheet.IsCreated {
			timesheet.MakeNewID()
			newTimesheetEntities = append(newTimesheetEntities, timesheet.ToEntity())
		}

		newTimesheetLessonHoursEntities = append(newTimesheetLessonHoursEntities,
			dto.ListTimesheetLessonHours(timesheet.GetTimesheetLessonHoursNew()).ToEntities()...)
	}

	return newTimesheetEntities, newTimesheetLessonHoursEntities
}

func buildRemoveTimesheets(timesheets []*dto.Timesheet) ([]string, []*entity.TimesheetLessonHours) {
	var (
		removeTimesheetIDs                 []string
		removeTimesheetLessonHoursEntities []*entity.TimesheetLessonHours
	)
	for _, timesheet := range timesheets {
		if timesheet.IsDeleted {
			removeTimesheetIDs = append(removeTimesheetIDs, timesheet.ID)
		}
		for _, timesheetLessonHours := range timesheet.ListTimesheetLessonHours {
			if timesheetLessonHours.IsDeleted {
				removeTimesheetLessonHoursEntities = append(removeTimesheetLessonHoursEntities, timesheetLessonHours.ToEntity())
			}
		}
	}

	return removeTimesheetIDs, removeTimesheetLessonHoursEntities
}

func (s *AutoCreateTimesheetServiceImpl) buildTransportExpense(ctx context.Context, newTS []*entity.Timesheet) (newTEs []*entity.TransportationExpense, err error) {
	if len(newTS) == 0 {
		return nil, nil
	}

	// just have 1 location for every create so just need get location ID at first Item
	locationID := newTS[0].LocationID.String

	// get staff in timesheet
	mapStaffs := make(map[string]struct{})
	for _, ts := range newTS {
		mapStaffs[ts.StaffID.String] = struct{}{}
	}
	staffIDs := make([]string, 0, len(mapStaffs))
	for key := range mapStaffs {
		staffIDs = append(staffIDs, key)
	}

	mapStaffTEs, err := s.StaffTransportationExpenseRepo.FindListTransportExpensesByStaffIDsAndLocation(ctx, s.DB, staffIDs, locationID)
	if err != nil {
		return nil, err
	}

	// calculate timesheet TEs
	for _, ts := range newTS {
		for _, staffTE := range mapStaffTEs[ts.StaffID.String] {
			te := entity.NewTransportExpenses()
			te.TimesheetID = ts.TimesheetID
			te.TransportationType = staffTE.TransportationType
			te.TransportationFrom = staffTE.TransportationFrom
			te.TransportationTo = staffTE.TransportationTo
			te.CostAmount = staffTE.CostAmount
			te.RoundTrip = staffTE.RoundTrip
			te.Remarks = staffTE.Remarks

			newTEs = append(newTEs, te)
		}
	}

	return newTEs, nil
}
