package timesheet

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/timesheet/infrastructure/repository"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) createDeletedLessonSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonDeletedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *bpb.EvtLesson_DeletedLessons_:
			stepState.DeletedLessonIDs = r.GetDeletedLessons().LessonIds
			stepState.FoundChanForJetStream <- r.Message
			return true, nil
		}
		return false, errors.New("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonDeleted, opts, handlerLessonDeletedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userDeleteALesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := lpb.DeleteLessonRequest{
		LessonId: stepState.CurrentLessonID,
	}
	ctx, err := s.createDeletedLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createDeletedLessonSubscription: %w", err)
	}
	_, stepState.ResponseErr = lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).DeleteLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createRecurringLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	loc := LoadLocalLocation()
	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(time.Date(2022, 7, 9, 9, 0, 0, 0, loc)),
		EndTime:         timestamppb.New(time.Date(2022, 7, 9, 10, 0, 0, 0, loc)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      stepState.CenterIDs[rand.Intn(len(stepState.CenterIDs))],
		StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*lpb.Material{},
		SavingOption: &lpb.CreateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
			Recurrence: &lpb.Recurrence{
				EndDate: timestamppb.New(time.Date(2022, 7, 31, 10, 0, 0, 0, loc)),
			},
		},
		SchedulingStatus: lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
	}
	addedStudentIDs := make(map[string]bool)
	for i := 0; i < len(stepState.StudentIDWithCourseID); i += 2 {
		studentID := stepState.StudentIDWithCourseID[i]
		courseID := stepState.StudentIDWithCourseID[i+1]
		if _, ok := addedStudentIDs[studentID]; ok {
			continue
		}
		addedStudentIDs[studentID] = true
		req.StudentInfoList = append(req.StudentInfoList, &lpb.CreateLessonRequest_StudentInfo{
			StudentId:        studentID,
			CourseId:         courseID,
			AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ATTEND,
			AttendanceNotice: lpb.StudentAttendanceNotice_NOTICE_EMPTY,
			AttendanceReason: lpb.StudentAttendanceReason_REASON_EMPTY,
		})
	}
	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &lpb.Material{
			Resource: &lpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}

	stepState.Request = req
	res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).CreateLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.ResponseErr = err
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	stepState.Response = res
	stepState.CurrentLessonID = res.GetId()
	schedulerID, err := s.getSchedulerIDByLessonID(ctx, res.GetId())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.OldSchedulerID = schedulerID

	lessons, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	lessonIDs := make([]string, 0, len(lessons))

	for _, v := range lessons {
		lessonIDs = append(lessonIDs, v.LessonID)
	}
	stepState.LessonIDs = lessonIDs

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userDeleteLessonRecurring(ctx context.Context, lessonIndex string, method string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonIndexNum, err := strconv.Atoi(lessonIndex)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := lpb.DeleteLessonRequest{
		LessonId: stepState.LessonIDs[lessonIndexNum],
	}
	switch method {
	case "one_time":
		req.SavingOption = &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME}
	case "recurring":
		req.SavingOption = &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE}
	}
	ctx, err = s.createDeletedLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createDeletedLessonSubscription: %w", err)
	}
	_, stepState.ResponseErr = lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).DeleteLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) getSchedulerIDByLessonID(ctx context.Context, lessonID string) (string, error) {
	query := `select scheduler_id from lessons where lesson_id = $1`
	var schedulerID string
	row := s.BobDB.QueryRow(ctx, query, lessonID)
	if err := row.Scan(&schedulerID); err != nil {
		return "", err
	}
	return schedulerID, nil
}

func (s *Suite) retrieveLessonChainByLessonID(ctx context.Context) ([]*domain.Lesson, error) {
	stepState := StepStateFromContext(ctx)
	query := `select lesson_id from lessons where scheduler_id = (select scheduler_id from lessons where lesson_id = $1) order by start_time asc`
	rows, err := s.BobDB.Query(ctx, query, stepState.CurrentLessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lessons []*domain.Lesson
	lessonRepo := repo.LessonRepo{}
	var lessonID string
	for rows.Next() {
		err = rows.Scan(&lessonID)
		if err != nil {
			return nil, err
		}
		lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, lessonID)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	return lessons, nil
}

func (s *Suite) lockLesson(ctx context.Context, isLock string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.CurrentLessonID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stepState.CurrentLessonID is empty")
	}
	isLockStr := isLock == "true"
	sql := fmt.Sprintf("UPDATE lessons SET is_locked = %s where lesson_id = $1 and deleted_at is null", strconv.FormatBool(isLockStr))
	_, err := s.BobDB.Exec(ctx, sql, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lockLessons(ctx context.Context, lockedLessons string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessons := getLessonByLessonListStr(ctx, lockedLessons)

	for _, v := range lessons {
		stepState.CurrentLessonID = v
		ctx, err := s.lockLesson(ctx, "true")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) addOtherWorkingHoursToTimesheet(ctx context.Context, otherWorkingHoursRecords string) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var timesheetID string
	numOfOtherWorkingHoursRecord, err := strconv.Atoi(otherWorkingHoursRecords)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stmt := `SELECT timesheet_id FROM timesheet_lesson_hours WHERE lesson_id = $1 AND deleted_at IS NULL`
	err = s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentLessonID).Scan(&timesheetID)
	if err != nil {
		return ctx, fmt.Errorf("addOtherWorkingHoursToTimesheet got error: %v", err)
	}
	for i := 0; i < numOfOtherWorkingHoursRecord; i++ {
		_, err = initOtherWorkingHours(ctx, timesheetID, initTimesheetConfigID1, time.Now(), strconv.Itoa(constants.ManabieSchool))
		if err != nil {
			return nil, err
		}
	}
	stepState.NumberOfOtherWorkingHours = int32(numOfOtherWorkingHoursRecord)
	stepState.CurrentTimesheetIDs = []string{timesheetID}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) addOtherWorkingHoursToTimesheetLessonRecurring(ctx context.Context, otherWorkingHoursRecords string) (context.Context, error) {
	time.Sleep(5 * time.Second)
	stepState := StepStateFromContext(ctx)

	numOfOtherWorkingHoursRecord, err := strconv.Atoi(otherWorkingHoursRecords)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	timesheetLessonHoursRepo := &repository.TimesheetLessonHoursRepoImpl{}
	timesheetLessonHoursEntities, err := timesheetLessonHoursRepo.FindTimesheetLessonHoursByLessonIDs(ctx, s.TimesheetDB, stepState.LessonIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("get TimesheetLessonHoursByLessonIDs error: %v", err)
	}
	timesheetIDs := make([]string, 0, len(timesheetLessonHoursEntities))
	for _, v := range timesheetLessonHoursEntities {
		timesheetIDs = append(timesheetIDs, v.TimesheetID.String)
	}
	for _, timesheetID := range timesheetIDs {
		for i := 0; i < numOfOtherWorkingHoursRecord; i++ {
			_, err = initOtherWorkingHours(ctx, timesheetID, initTimesheetConfigID1, time.Now(), strconv.Itoa(constants.ManabieSchool))
			if err != nil {
				return nil, err
			}
		}
	}
	stepState.CurrentTimesheetIDs = timesheetIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) addTransportationExpenseToTimesheet(ctx context.Context, transportationExpenseRecords string) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var timesheetID string
	numOfTransportationExpenseRecord, err := strconv.Atoi(transportationExpenseRecords)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stmt := `SELECT timesheet_id FROM timesheet_lesson_hours WHERE lesson_id = $1 AND deleted_at IS NULL`
	err = s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentLessonID).Scan(&timesheetID)
	if err != nil {
		return ctx, fmt.Errorf("addTransportationExpenseToTimesheet got error: %v", err)
	}
	for i := 0; i < numOfTransportationExpenseRecord; i++ {
		_, err = initTransportExpenses(ctx, timesheetID, strconv.Itoa(constants.ManabieSchool))
		if err != nil {
			return nil, err
		}
	}
	stepState.CurrentTimesheetIDs = []string{timesheetID}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) addTransportationExpenseToTimesheetLessonRecurring(ctx context.Context, transportationExpenseRecords string) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var timesheetIDs []string
	numOfTransportationExpenseRecord, err := strconv.Atoi(transportationExpenseRecords)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stmtTimesheetLessonHours := `SELECT timesheet_id FROM timesheet_lesson_hours WHERE lesson_id = ANY($1::_TEXT) AND deleted_at IS NULL`
	rows, err := s.TimesheetDB.Query(ctx, stmtTimesheetLessonHours, database.TextArray(stepState.LessonIDs))
	defer rows.Close()
	if err != nil {
		return ctx, fmt.Errorf("query getTimesheetLessonHours got error: %v", err)
	}
	for rows.Next() {
		var timesheetID string
		if err = rows.Scan(&timesheetID); err != nil {
			return ctx, fmt.Errorf("get value of row got error: %v", err)
		}
		timesheetIDs = append(timesheetIDs, timesheetID)
	}
	for _, timesheetID := range timesheetIDs {
		for i := 0; i < numOfTransportationExpenseRecord; i++ {
			_, err = initTransportExpenses(ctx, timesheetID, strconv.Itoa(constants.ManabieSchool))
			if err != nil {
				return nil, err
			}
		}
	}
	stepState.CurrentTimesheetIDs = timesheetIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkTimesheetLessonHoursIsDeleted(ctx context.Context, total int, isDeleted string) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	timesheetLessonHoursRepo := &repository.TimesheetLessonHoursRepoImpl{}
	timesheetLessonHours, err := timesheetLessonHoursRepo.FindTimesheetLessonHoursByLessonIDs(ctx, s.TimesheetDB, stepState.DeletedLessonIDs)
	if err != nil {
		return ctx, err
	}
	if isDeleted == "deleted" && len(timesheetLessonHours) == total-1 {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d timesheetLessonHours be deleted but got %d still existed", total, len(timesheetLessonHours))
}

func (s *Suite) checkTimesheetLessonHoursRecurringLessonBeDeleted(ctx context.Context, total int, isDeleted string) (context.Context, error) {
	time.Sleep(5 * time.Second)
	stepState := StepStateFromContext(ctx)
	var timesheetLessonHoursCount int
	var timesheetIDs []string
	stmtTimesheetLessonHours := `SELECT count(timesheet_id) FROM timesheet_lesson_hours WHERE lesson_id = ANY($1::_TEXT) AND deleted_at IS NOT NULL`
	err := s.TimesheetDB.QueryRow(ctx, stmtTimesheetLessonHours, database.TextArray(stepState.LessonIDs)).Scan(&timesheetLessonHoursCount)
	if err != nil {
		return ctx, fmt.Errorf("query TimesheetLessonHours got error: %v", err)
	}
	if isDeleted == "deleted" && timesheetLessonHoursCount == total {
		stmtTimesheetLessonHours = `SELECT timesheet_id FROM timesheet_lesson_hours WHERE lesson_id = ANY($1::_TEXT) AND deleted_at IS NOT NULL`
		rows, err := s.TimesheetDB.Query(ctx, stmtTimesheetLessonHours, database.TextArray(stepState.LessonIDs))
		defer rows.Close()
		if err != nil {
			return ctx, fmt.Errorf("query getTimesheetLessonHours got error: %v", err)
		}
		for rows.Next() {
			var timesheetID string
			if err = rows.Scan(&timesheetID); err != nil {
				return ctx, fmt.Errorf("get value of row got error: %v", err)
			}
			timesheetIDs = append(timesheetIDs, timesheetID)
		}
		stepState.CurrentTimesheetIDs = timesheetIDs
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf(
		"expect %d TimesheetLessonHours be deleted but %d deleted",
		total, timesheetLessonHoursCount)
}

func (s *Suite) checkTimesheetRecurringLessonBeDeleted(ctx context.Context, total int) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var timesheetCount int
	var timesheetIDs []string
	stmtTimesheetLessonHours := `SELECT timesheet_id FROM timesheet_lesson_hours WHERE lesson_id = ANY($1::_TEXT) AND deleted_at IS NOT NULL`
	rows, err := s.TimesheetDB.Query(ctx, stmtTimesheetLessonHours, database.TextArray(stepState.LessonIDs))
	defer rows.Close()
	if err != nil {
		return ctx, fmt.Errorf("query getTimesheetLessonHours got error: %v", err)
	}
	for rows.Next() {
		var timesheetID string
		if err = rows.Scan(&timesheetID); err != nil {
			return ctx, fmt.Errorf("get value of row got error: %v", err)
		}
		timesheetIDs = append(timesheetIDs, timesheetID)
	}
	stmtTimesheet := `SELECT count(timesheet_id) FROM timesheet WHERE timesheet_id = ANY($1::_TEXT) AND deleted_at IS NOT NULL`
	err = s.TimesheetDB.QueryRow(ctx, stmtTimesheet, database.TextArray(timesheetIDs)).Scan(&timesheetCount)
	if err != nil {
		return ctx, fmt.Errorf("query get list of timesheet got error:%v", err)
	}
	if timesheetCount == total {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf(
		"expect %v of timesheet will be deleted but only %v / %v be deleted", total, timesheetCount, total)
}

func (s *Suite) checkTimesheetRecurringLessonNotDeleted(ctx context.Context, total int) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var timesheetCount int
	var timesheetNoDeletedLessonHoursCount int
	stmtTimesheet := `SELECT count(timesheet_id) FROM timesheet WHERE timesheet_id = ANY($1::_TEXT) AND deleted_at IS NULL`
	err := s.TimesheetDB.QueryRow(ctx, stmtTimesheet, database.TextArray(stepState.CurrentTimesheetIDs)).Scan(&timesheetCount)
	if err != nil {
		return ctx, fmt.Errorf("query Timesheet got error: %v", err)
	}
	stmtTimesheetLessonHours := `SELECT count(timesheet_id) FROM timesheet_lesson_hours WHERE lesson_id = ANY($1::_TEXT) AND deleted_at IS NULL`
	err = s.TimesheetDB.QueryRow(ctx, stmtTimesheetLessonHours, database.TextArray(stepState.LessonIDs)).Scan(&timesheetNoDeletedLessonHoursCount)
	if err != nil {
		return ctx, fmt.Errorf("query TimesheetLessonHours got error: %v", err)
	}
	timesheetCount += timesheetNoDeletedLessonHoursCount
	if timesheetCount == total {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v timesheet not deleted but %v deleted", total, timesheetCount)
}

func (s *Suite) checkTimesheetLessonHoursRemaining(ctx context.Context, total int) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var timesheetLessonHoursRemainingCount int
	stmtTimesheetLessonHours := `SELECT count(timesheet_id) FROM timesheet_lesson_hours WHERE lesson_id = ANY($1::_TEXT) AND deleted_at IS NULL`
	err := s.TimesheetDB.QueryRow(ctx, stmtTimesheetLessonHours, database.TextArray(stepState.LessonIDs)).Scan(&timesheetLessonHoursRemainingCount)
	if err != nil {
		return ctx, fmt.Errorf("query checkTimesheetLessonHoursRemaining got error: %v", err)
	}
	if timesheetLessonHoursRemainingCount == total {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf(
		"expect %d TimesheetLessonHours be existed but got %d", total, timesheetLessonHoursRemainingCount)
}

func (s *Suite) checkTimesheetIsDeleted(ctx context.Context) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var count int
	var timesheetID string
	stmtTimesheetLessonHours := `SELECT timesheet_id FROM timesheet_lesson_hours WHERE lesson_id = $1 AND deleted_at IS NOT NULL`
	err := s.TimesheetDB.QueryRow(ctx, stmtTimesheetLessonHours, stepState.DeletedLessonIDs[0]).Scan(&timesheetID)
	if err != nil {
		return ctx, fmt.Errorf("query TimesheetLessonHours got error: %v", err)
	}
	stmtTimesheet := `SELECT count(timesheet_id) FROM timesheet WHERE timesheet_id = $1 AND deleted_at IS NOT NULL`
	err = s.TimesheetDB.QueryRow(ctx, stmtTimesheet, timesheetID).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("query Timesheet got error: %v", err)
	}
	if count != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("checkTimesheetIsDeleted unexpected %d -> timesheet %s", count, timesheetID)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkTimesheetIsNotDeleted(ctx context.Context) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	var count int
	var timesheetID string
	stmtTimesheetLessonHours := `SELECT timesheet_id FROM timesheet_lesson_hours WHERE lesson_id = $1 AND deleted_at IS NOT NULL`
	err := s.TimesheetDB.QueryRow(ctx, stmtTimesheetLessonHours, stepState.DeletedLessonIDs[0]).Scan(&timesheetID)
	if err != nil {
		return ctx, fmt.Errorf("query TimesheetLessonHours got error: %v", err)
	}
	stmtTimesheet := `SELECT count(timesheet_id) FROM timesheet WHERE timesheet_id = $1 AND deleted_at IS NULL`
	err = s.TimesheetDB.QueryRow(ctx, stmtTimesheet, timesheetID).Scan(&count)
	if err != nil {
		return ctx, fmt.Errorf("query Timesheet got error: %v", err)
	}
	if count != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("checkTimesheetIsNotDeleted unexpected %d -> timesheet %s", count, timesheetID)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkOtherWorkingHoursStillExisted(ctx context.Context, total int) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	otherWorkingHoursRepo := &repository.OtherWorkingHoursRepoImpl{}
	otherWorkingHours, err := otherWorkingHoursRepo.FindListOtherWorkingHoursByTimesheetIDs(ctx, s.TimesheetDB, database.TextArray(stepState.CurrentTimesheetIDs))
	if err != nil {
		return ctx, err
	}
	if len(otherWorkingHours) == total {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d OtherWorkingHours existed but got %v", total, len(otherWorkingHours))
}

func (s *Suite) checkTransportationExpensesStillExisted(ctx context.Context, total int) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	transportationExpensesRepo := &repository.TransportationExpenseRepoImpl{}
	transportationExpenses, err := transportationExpensesRepo.FindListTransportExpensesByTimesheetIDs(ctx, s.TimesheetDB, database.TextArray(stepState.CurrentTimesheetIDs))
	if err != nil {
		return ctx, err
	}
	if len(transportationExpenses) == total {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("expect %d TransportationExpenses existed but got %v", total, len(transportationExpenses))
}

func (s *Suite) updateLessonRequestDefault(ctx context.Context) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	createLessonReq, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected *lpb.CreateLessonRequest, but got %T type", stepState.Request.(*lpb.CreateLessonRequest))
	}
	lessonChain, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(lessonChain) < 2 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of lesson in chain is not correct")
	}
	lessonC := lessonChain[2]
	stepState.OldEndDate = lessonChain[1].EndTime
	updateLessonReq := &bpb.UpdateLessonRequest{
		LessonId:        lessonC.LessonID,
		StartTime:       timestamppb.New(lessonC.StartTime),
		EndTime:         timestamppb.New(lessonC.EndTime),
		TeachingMedium:  createLessonReq.TeachingMedium,
		TeachingMethod:  createLessonReq.TeachingMethod,
		TeacherIds:      createLessonReq.TeacherIds,
		CenterId:        createLessonReq.LocationId,
		StudentInfoList: []*bpb.UpdateLessonRequest_StudentInfo{},
		Materials:       []*bpb.Material{},
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
			Recurrence: &bpb.Recurrence{
				EndDate: createLessonReq.SavingOption.Recurrence.EndDate,
			},
		},
	}
	for _, v := range createLessonReq.StudentInfoList {
		updateLessonReq.StudentInfoList = append(updateLessonReq.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        v.StudentId,
			CourseId:         v.CourseId,
			LocationId:       v.LocationId,
			AttendanceStatus: bpb.StudentAttendStatus(v.AttendanceStatus),
		})
	}
	materials := make([]*bpb.Material, 0, len(createLessonReq.Materials))
	for _, v := range createLessonReq.Materials {
		switch resource := v.Resource.(type) {
		case *lpb.Material_BrightcoveVideo_:
			material := &bpb.Material{
				Resource: &bpb.Material_BrightcoveVideo_{
					BrightcoveVideo: &bpb.Material_BrightcoveVideo{
						Name: resource.BrightcoveVideo.Name,
						Url:  resource.BrightcoveVideo.Url,
					}}}
			materials = append(materials, material)
		case *lpb.Material_MediaId:
			material := &bpb.Material{
				Resource: &bpb.Material_MediaId{
					MediaId: resource.MediaId,
				}}
			materials = append(materials, material)
		default:
			return nil, status.Error(codes.Internal, fmt.Errorf(`unexpected material's type %T`, resource).Error())
		}
	}
	updateLessonReq.Materials = materials
	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userChangedEndDate(ctx context.Context, _expectedTypeOfDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	endDate := genDateTimeFormat(_expectedTypeOfDate)
	ctx, err := s.updateLessonRequestDefault(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq, ok := stepState.Request.(*bpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	updateLessonReq.SavingOption = &bpb.UpdateLessonRequest_SavingOption{
		Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
		Recurrence: &bpb.Recurrence{
			EndDate: timestamppb.New(endDate),
		},
	}
	stepState.Request = updateLessonReq
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateLessonBySavingWeekly(ctx context.Context) (context.Context, error) {
	time.Sleep(10 * time.Second)
	stepState := StepStateFromContext(ctx)
	updatedReq := stepState.Request.(*bpb.UpdateLessonRequest)
	ctx, err := s.createDeletedLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	stepState.Response, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.BobConn).UpdateLesson(s.CommonSuite.SignedCtx(ctx), updatedReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	stepState.CurrentLessonID = updatedReq.LessonId
	return StepStateToContext(ctx, stepState), nil
}

func getLessonByLessonListStr(ctx context.Context, lessonListStr string) []string {
	stepState := StepStateFromContext(ctx)
	lessons := make([]string, 0)

	for i := 0; i < len(stepState.LessonIDs); i++ {
		for _, y := range strings.Split(lessonListStr, ",") {
			if fmt.Sprint(i) == y {
				lessons = append(lessons, stepState.LessonIDs[i])
			}
		}
	}
	return lessons
}
