package timesheet

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/jackc/pgtype"
)

type GetTimesheetServiceImpl struct {
	DB database.Ext

	TimesheetRepo interface {
		FindTimesheetByTimesheetArgs(ctx context.Context, db database.QueryExecer, timesheetArgs *dto.TimesheetQueryArgs) ([]*entity.Timesheet, error)
		FindTimesheetByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) ([]*entity.Timesheet, error)
		FindTimesheetByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*entity.Timesheet, error)
	}

	TimesheetLessonHoursRepo interface {
		FindByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) ([]*entity.TimesheetLessonHours, error)
		FindTimesheetLessonHoursByLessonIDs(ctx context.Context, db database.QueryExecer, lessonIDs []string) ([]*entity.TimesheetLessonHours, error)
	}

	OtherWorkingHoursRepo interface {
		FindListOtherWorkingHoursByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs pgtype.TextArray) ([]*entity.OtherWorkingHours, error)
	}

	TransportationExpenseRepo interface {
		FindListTransportExpensesByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs pgtype.TextArray) ([]*entity.TransportationExpense, error)
	}
}

func (s *GetTimesheetServiceImpl) GetTimesheetByLessonIDs(ctx context.Context, lessonIDs []string) ([]*dto.Timesheet, error) {
	timesheetEntities, err := s.TimesheetRepo.FindTimesheetByLessonIDs(ctx, s.DB, lessonIDs)
	if err != nil {
		return nil, fmt.Errorf("GetTimesheetByLessonIDs -> FindTimesheetByLessonIDs %v got error: %s", lessonIDs, err.Error())
	}

	var (
		timesheetIDs     = make([]string, 0, len(timesheetEntities))
		mapIDTimesheet   = map[string]*dto.Timesheet{}
		resultTimesheets = make([]*dto.Timesheet, 0, len(timesheetEntities))
	)
	for _, e := range timesheetEntities {
		timesheetIDs = append(timesheetIDs, dto.NewTimesheetFromEntity(e).ID)
		mapIDTimesheet[e.TimesheetID.String] = dto.NewTimesheetFromEntity(e)
	}
	if len(mapIDTimesheet) == 0 {
		return nil, nil
	}

	timesheetLessonHoursEntities, err := s.TimesheetLessonHoursRepo.FindByTimesheetIDs(ctx, s.DB, timesheetIDs)
	if err != nil {
		return nil, fmt.Errorf("GetTimesheetByLessonIDs -> FindByTimesheetIDs %v got error: %s", timesheetIDs, err.Error())
	}
	appendTimesheetLessonHours(mapIDTimesheet, timesheetLessonHoursEntities)

	otherWorkingHoursEntities, err := s.OtherWorkingHoursRepo.FindListOtherWorkingHoursByTimesheetIDs(ctx, s.DB, database.TextArray(timesheetIDs))
	if err != nil {
		return nil, fmt.Errorf("GetTimesheetByLessonIDs -> FindListOtherWorkingHoursByTimesheetIDs %v got error: %s", timesheetIDs, err.Error())
	}
	appendOtherWorkingHours(mapIDTimesheet, otherWorkingHoursEntities)

	transportationExpenseEntities, err := s.TransportationExpenseRepo.FindListTransportExpensesByTimesheetIDs(ctx, s.DB, database.TextArray(timesheetIDs))
	if err != nil {
		return nil, fmt.Errorf("GetTimesheetByLessonIDs -> FindListTransportExpensesByTimesheetIDs %v got error: %s", timesheetIDs, err.Error())
	}
	appendTransportationExpenses(mapIDTimesheet, transportationExpenseEntities)

	for _, timesheet := range mapIDTimesheet {
		resultTimesheets = append(resultTimesheets, timesheet)
	}
	return resultTimesheets, nil
}

func (s *GetTimesheetServiceImpl) GetTimesheet(ctx context.Context, timesheetQueryArgs *dto.TimesheetQueryArgs, timesheetQueryOptions *dto.TimesheetGetOptions) ([]*dto.Timesheet, error) {
	err := timesheetQueryArgs.Validate()
	if err != nil {
		return nil, err
	}
	timesheetQueryArgs.Normalize()
	timesheetEntities, err := s.TimesheetRepo.FindTimesheetByTimesheetArgs(ctx, s.DB, timesheetQueryArgs)
	if err != nil {
		return nil, err
	}

	if getTimesheetInfoOnly(timesheetQueryOptions) {
		timesheets := make([]*dto.Timesheet, 0, len(timesheetEntities))
		for _, e := range timesheetEntities {
			timesheets = append(timesheets, dto.NewTimesheetFromEntity(e))
		}
		return timesheets, nil
	}

	var (
		mapIDTimesheet   = map[string]*dto.Timesheet{}
		timesheetIDs     = make([]string, 0, len(timesheetEntities))
		resultTimesheets = make([]*dto.Timesheet, 0, len(timesheetEntities))
	)

	for _, e := range timesheetEntities {
		timesheet := dto.NewTimesheetFromEntity(e)
		mapIDTimesheet[timesheet.ID] = timesheet
		timesheetIDs = append(timesheetIDs, timesheet.ID)
	}

	if len(mapIDTimesheet) == 0 {
		return nil, nil
	}

	if timesheetQueryOptions.IsGetListTimesheetLessonHours {
		timesheetLessonHoursEntities, err := s.TimesheetLessonHoursRepo.FindByTimesheetIDs(ctx, s.DB, timesheetIDs)
		if err != nil {
			return nil, err
		}
		appendTimesheetLessonHours(mapIDTimesheet, timesheetLessonHoursEntities)
	}

	if timesheetQueryOptions.IsGetListOtherWorkingHours {
		otherWorkingHoursEntities, err := s.OtherWorkingHoursRepo.FindListOtherWorkingHoursByTimesheetIDs(ctx, s.DB, database.TextArray(timesheetIDs))
		if err != nil {
			return nil, err
		}
		appendOtherWorkingHours(mapIDTimesheet, otherWorkingHoursEntities)
	}

	if timesheetQueryOptions.IsGetListTransportationExpense {
		transportationExpenseEntities, err := s.TransportationExpenseRepo.FindListTransportExpensesByTimesheetIDs(ctx, s.DB, database.TextArray(timesheetIDs))
		if err != nil {
			return nil, err
		}
		appendTransportationExpenses(mapIDTimesheet, transportationExpenseEntities)
	}

	for _, timesheet := range mapIDTimesheet {
		resultTimesheets = append(resultTimesheets, timesheet)
	}

	return resultTimesheets, nil
}

func getTimesheetInfoOnly(timesheetGetOptions *dto.TimesheetGetOptions) bool {
	if (!timesheetGetOptions.IsGetListTimesheetLessonHours) &&
		(!timesheetGetOptions.IsGetListOtherWorkingHours) &&
		(!timesheetGetOptions.IsGetListTransportationExpense) {
		return true
	}
	return false
}

func appendOtherWorkingHours(mapIDTimesheet map[string]*dto.Timesheet, otherWorkingHoursEntities []*entity.OtherWorkingHours) {
	for _, e := range otherWorkingHoursEntities {
		otherWorkingHours := dto.NewOtherWorkingHoursFromEntity(e)
		timesheet := mapIDTimesheet[otherWorkingHours.TimesheetID]
		timesheet.ListOtherWorkingHours = append(timesheet.ListOtherWorkingHours, otherWorkingHours)
	}
}

func appendTimesheetLessonHours(mapIDTimesheet map[string]*dto.Timesheet, timesheetLessonHoursEntities []*entity.TimesheetLessonHours) {
	for _, e := range timesheetLessonHoursEntities {
		timesheetLessonHours := dto.NewTimesheetLessonHoursFromEntity(e)
		timesheet := mapIDTimesheet[timesheetLessonHours.TimesheetID]
		timesheet.ListTimesheetLessonHours = append(timesheet.ListTimesheetLessonHours, timesheetLessonHours)
	}
}

func appendTransportationExpenses(mapIDTimesheet map[string]*dto.Timesheet, transportationExpensesEntities []*entity.TransportationExpense) {
	for _, e := range transportationExpensesEntities {
		transportationExpense := dto.NewTransportExpensesFromEntity(e)
		timesheet := mapIDTimesheet[transportationExpense.TimesheetID]
		timesheet.ListTransportationExpenses = append(timesheet.ListTransportationExpenses, transportationExpense)
	}
}
