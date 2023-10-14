package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AutoCreateTimesheetFlagServiceImpl struct {
	DB database.Ext

	AutoCreateFlagRepo interface {
		Upsert(ctx context.Context, db database.QueryExecer, flagE *entity.AutoCreateTimesheetFlag) error
		FindAutoCreatedFlagByStaffID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.AutoCreateTimesheetFlag, error)
	}

	AutoCreateFlagLogRepo interface {
		InsertFlagLog(ctx context.Context, db database.QueryExecer, flagE *entity.AutoCreateFlagActivityLog) (*entity.AutoCreateFlagActivityLog, error)
		GetAutoCreateFlagActivityLogByStaffIDs(ctx context.Context, db database.QueryExecer, lessonStartDate time.Time, teacherIDs []string) ([]*entity.AutoCreateFlagActivityLog, error)
		SoftDeleteFlagLogsAfterTime(ctx context.Context, db database.QueryExecer, staffID string, startTime time.Time) error
	}

	TimesheetRepo interface {
		GetStaffTimesheetIDsAfterDateCanChange(ctx context.Context, db database.QueryExecer, staffID string, date time.Time) ([]string, error)
		RemoveTimesheetRemarkByTimesheetIDs(ctx context.Context, db database.QueryExecer, ids []string) error
	}

	TimesheetLessonHoursRepo interface {
		UpdateAutoCreateFlagStateAfterTime(ctx context.Context, db database.QueryExecer, timesheetIDs []string, updateTime time.Time, flagOn bool) error
		MapExistingLessonHoursByTimesheetIds(ctx context.Context, db database.QueryExecer, ids []string) (map[string]struct{}, error)
	}

	OtherWorkingHoursRepo interface {
		MapExistingOWHsByTimesheetIds(ctx context.Context, db database.QueryExecer, ids []string) (map[string]struct{}, error)
	}
}

func (s *AutoCreateTimesheetFlagServiceImpl) UpsertFlag(ctx context.Context, req *dto.AutoCreateTimesheetFlag) error {
	flagLogDto := dto.AutoCreateFlagActivityLog{
		ID:      idutil.ULIDNow(),
		StaffID: req.StaffID,
		FlagOn:  req.FlagOn,
	}

	now := time.Now().In(timeutil.Timezone(pbc.COUNTRY_JP)) // date in Japan timezone
	dateNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Update flag infos
		if err := s.AutoCreateFlagRepo.Upsert(ctx, tx, req.ToEntity()); err != nil {
			return fmt.Errorf("create or update auto create flag error: %v", err)
		}

		// soft delete flag logs within the day before inserting the new flag log, this will retain one flag log per day
		err := s.AutoCreateFlagLogRepo.SoftDeleteFlagLogsAfterTime(ctx, tx, req.StaffID, dateNow)
		if err != nil {
			return fmt.Errorf("soft delete flag logs within time range error: %v", err)
		}

		_, err = s.AutoCreateFlagLogRepo.InsertFlagLog(ctx, tx, flagLogDto.ToEntity())
		if err != nil {
			return fmt.Errorf("create log for auto create flag error: %v", err)
		}

		return nil
	})
	if err != nil {
		return status.Errorf(codes.Internal, "transaction error: %v", err)
	}

	return nil
}

func (s *AutoCreateTimesheetFlagServiceImpl) UpdateLessonHoursFlag(ctx context.Context, req *dto.AutoCreateTimesheetFlag) error {
	timeNow := time.Now().In(timeutil.Timezone(pbc.COUNTRY_JP)) // date in Japan timezone
	dateNow := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())

	timesheetsIDs, err := s.TimesheetRepo.GetStaffTimesheetIDsAfterDateCanChange(ctx, s.DB, req.StaffID, dateNow)
	if err != nil {
		return status.Errorf(codes.Internal, "get timesheet ids after date error: %v", err)
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// update future timesheet lesson hours flag
		if len(timesheetsIDs) > 0 {
			err = s.TimesheetLessonHoursRepo.UpdateAutoCreateFlagStateAfterTime(ctx, tx, timesheetsIDs, dateNow, req.FlagOn)
			if err != nil {
				return fmt.Errorf("update future auto create flag error: %v", err)
			}

			if !req.FlagOn {
				mapTSLessonHours, err := s.TimesheetLessonHoursRepo.MapExistingLessonHoursByTimesheetIds(ctx, tx, timesheetsIDs)
				if err != nil {
					return fmt.Errorf("get list timesheet lesson hours by timesheet ids error: %v", err)
				}
				mapTSOWHs, err := s.OtherWorkingHoursRepo.MapExistingOWHsByTimesheetIds(ctx, tx, timesheetsIDs)
				if err != nil {
					return fmt.Errorf("get list timesheet OHWs by timesheet ids error: %v", err)
				}

				removeRemarkIDs := []string{}
				for _, id := range timesheetsIDs {
					_, found1 := mapTSLessonHours[id]
					_, found2 := mapTSOWHs[id]

					if !found1 && !found2 {
						removeRemarkIDs = append(removeRemarkIDs, id)
					}
				}

				err = s.TimesheetRepo.RemoveTimesheetRemarkByTimesheetIDs(ctx, tx, removeRemarkIDs)
				if err != nil {
					return fmt.Errorf("remove remark for empty timesheet error: %v", err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return status.Errorf(codes.Internal, "transaction error: %v", err)
	}

	return nil
}

func (s *AutoCreateTimesheetFlagServiceImpl) GetAutoCreateFlagLogByTeacherIDs(ctx context.Context, lessonStartDate time.Time, teacherIDs []string) ([]*dto.AutoCreateFlagActivityLog, error) {
	autoCreateFlagActivityLogEntities, err := s.AutoCreateFlagLogRepo.GetAutoCreateFlagActivityLogByStaffIDs(ctx, s.DB, lessonStartDate, teacherIDs)
	if err != nil {
		return nil, fmt.Errorf("AutoCreateTimesheetFlagService.GetAutoCreateFlagActivityLogByStaffIDs error: %v", err.Error())
	}

	listAutoCreateFlagActivityLog := make([]*dto.AutoCreateFlagActivityLog, 0, len(autoCreateFlagActivityLogEntities))
	for _, activityLogE := range autoCreateFlagActivityLogEntities {
		listAutoCreateFlagActivityLog = append(listAutoCreateFlagActivityLog, dto.NewAutoCreateFlagActivityLogFromEntity(activityLogE))
	}

	return listAutoCreateFlagActivityLog, nil
}
