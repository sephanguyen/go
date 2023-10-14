package timesheet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/infrastructure/repository"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) buildUpsertAutoCreateTimesheetRequest(ctx context.Context, flagOnStatusStr string) (context.Context, error) {
	var (
		staffID      string
		err          error
		stepState    = StepStateFromContext(ctx)
		flagOnStatus = false
	)

	if flagOnStatusStr == "true" {
		flagOnStatus = true
	}

	staffID, err = getOneStaffIDInDB(ctx, stepState.CurrentUserID, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = &pb.UpdateAutoCreateTimesheetFlagRequest{
		StaffId: staffID,
		FlagOn:  flagOnStatus,
	}
	stepState.CurrentStaffID = staffID

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpsertAutoCreateTimesheetFlag(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Request == nil {
		stepState.Request = &pb.UpdateAutoCreateTimesheetFlagRequest{}
	}
	stepState.Response, stepState.ResponseErr =
		pb.NewAutoCreateTimesheetServiceClient(s.TimesheetConn).UpdateAutoCreateTimesheetFlag(contextWithToken(ctx), stepState.Request.(*pb.UpdateAutoCreateTimesheetFlagRequest))

	if stepState.ResponseErr != nil {
		stepState.StaffID = stepState.
			Request.(*pb.UpdateAutoCreateTimesheetFlagRequest).StaffId
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) userUpsertAutoCreateTimesheetFlagForTeachers(ctx context.Context, autoCreateValue string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var flagValue bool

	if autoCreateValue == "on" {
		flagValue = true
	}

	for _, teacherID := range stepState.TeacherIDs {
		stepState.Request = &pb.UpdateAutoCreateTimesheetFlagRequest{
			StaffId: teacherID,
			FlagOn:  flagValue,
		}
		stepState.Response, stepState.ResponseErr =
			pb.NewAutoCreateTimesheetServiceClient(s.TimesheetConn).UpdateAutoCreateTimesheetFlag(contextWithToken(ctx), stepState.Request.(*pb.UpdateAutoCreateTimesheetFlagRequest))
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) newUpsertAutoCreateTimesheetFlagData(ctx context.Context, flagOnStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.buildUpsertAutoCreateTimesheetRequest(ctx, flagOnStatus)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) verifyFlagStatusAfterUpsert(ctx context.Context, flagOnStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if stepState.Response != nil {
		if !stepState.Response.(*pb.UpdateAutoCreateTimesheetFlagResponse).Successful {
			return ctx, fmt.Errorf("error cannot upsert auto create timesheet flag record")
		}

		err := s.checkFlagRecord(ctx, flagOnStatus)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkFlagRecord(ctx context.Context, flagOnStatus string) error {
	stepState := StepStateFromContext(ctx)
	var count int

	stmt := `
		SELECT
			count(staff_id)
		FROM
			auto_create_timesheet_flag
		WHERE
			staff_id = $1
		AND
			flag_on  = $2
		AND
			deleted_at IS NOT NULL
		`
	err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentStaffID, flagOnStatus).Scan(&count)
	if err != nil {
		return err
	}

	return nil
}

func (s *Suite) adminUpdateTeacherAutoCreateFlag(ctx context.Context, flag string) (context.Context, error) {
	var (
		stepState = StepStateFromContext(ctx)
		flagOn    = false
	)

	if flag == "on" {
		flagOn = true
	}

	// get teacher ID
	if len(stepState.TeacherIDs) != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("teacher IDs must be equal to one")
	}

	request := &pb.UpdateAutoCreateTimesheetFlagRequest{
		StaffId: stepState.TeacherIDs[0],
		FlagOn:  flagOn,
	}

	// Change flag to state
	stepState.Response, stepState.ResponseErr = pb.NewAutoCreateTimesheetServiceClient(s.TimesheetConn).
		UpdateAutoCreateTimesheetFlag(contextWithToken(ctx), request)

	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserCreateALessonWithFutureDateInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"

	timeNow := time.Now().In(timeutil.Timezone(pbc.COUNTRY_JP)) // date in Japan timezone
	dateNow := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day()+5 /*5 day in future*/, 0, 0, 0, 0, timeNow.Location())

	ctx, _ = s.CommonSuite.UserCreateALessonInDateWithMissingFieldsInLessonmgmt(ctx, dateNow)
	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *Suite) checkTimesheetLessonHoursIsCreatedWithFlag(ctx context.Context, total int, isCreated string, flag string) (context.Context, error) {
	time.Sleep(5 * time.Second) // wait service process event from nats
	stepState := StepStateFromContext(ctx)

	timesheetLessonHoursRepo := &repository.TimesheetLessonHoursRepoImpl{}
	timesheetLessonHours, err := timesheetLessonHoursRepo.FindByTimesheetIDs(ctx, s.TimesheetDB, stepState.CurrentTimesheetIDs)
	if err != nil {
		return ctx, err
	}

	flagOn := false
	if flag == "on" {
		flagOn = true
	}

	if (isCreated == "created" && len(timesheetLessonHours) != total) ||
		(isCreated == "not created" && len(timesheetLessonHours) != total) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("timesheetLessonHours create not as expected, isCreated: %v, total: %d", isCreated, len(timesheetLessonHours))
	}

	for _, elm := range timesheetLessonHours {

		if elm.FlagOn.Bool != flagOn {
			return StepStateToContext(ctx, stepState), fmt.Errorf("timesheetLessonHours create not as expected, isCreated: %v, total: %d, flagOn:%t", isCreated, len(timesheetLessonHours), elm.FlagOn.Bool)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) flagInTimesheetLessonHoursChangeTo(ctx context.Context, flag string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timesheetLessonHoursRepo := &repository.TimesheetLessonHoursRepoImpl{}
	timesheetLessonHours, err := timesheetLessonHoursRepo.FindByTimesheetIDs(ctx, s.TimesheetDB, stepState.CurrentTimesheetIDs)
	if err != nil {
		return ctx, err
	}

	flagOn := false
	if flag == "on" {
		flagOn = true
	}

	for _, elm := range timesheetLessonHours {
		if elm.FlagOn.Bool != flagOn {
			return StepStateToContext(ctx, stepState), fmt.Errorf("timesheetLessonHours update not as expected, timesheetID:%s, flagOn:%t", elm.TimesheetID.String, elm.FlagOn.Bool)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) verifyFlagStatusAfterUpdate(ctx context.Context, flag string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	flagOn := "false"
	if flag == "on" {
		flagOn = "true"
	}

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if stepState.Response != nil {
		if !stepState.Response.(*pb.UpdateAutoCreateTimesheetFlagResponse).Successful {
			return ctx, fmt.Errorf("error cannot upsert auto create timesheet flag record")
		}

		err := s.checkFlagRecord(ctx, flagOn)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
