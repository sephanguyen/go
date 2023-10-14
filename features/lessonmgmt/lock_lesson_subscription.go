package lessonmgmt

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lessonmgmt_entities "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	user_entities "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/proto"
)

const (
	insertStaffStmtFormat               = `INSERT INTO staff (%s) VALUES (%s);`
	insertTimesheetStmtFormat           = `INSERT INTO timesheet (%s) VALUES (%s);`
	insertTimesheetLessonHourStmtFormat = `INSERT INTO timesheet_lesson_hours (%s) VALUES (%s);`
	insertLessonStmtFormat              = `INSERT INTO lessons (%s) VALUES (%s);`
	insertLessonTeacherStmtFormat       = `INSERT INTO lessons_teachers (%s) VALUES (%s);`
	insertTeacherStmtFormat             = `INSERT INTO teachers (%s) VALUES (%s);`
)

var locationIDs = []string{"1", "2", "3", "4", "5", "6", "7", "8"}

func (s *Suite) anExistingTimesheetForCurrentStaff(ctx context.Context, timesheetStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := initLocation(ctx, locationIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = initStaff(ctx, stepState.CurrentUserID, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = s.buildCreateTimesheeBasedOnStatus(ctx, timesheetStatus, "TODAY")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) buildCreateTimesheeBasedOnStatus(ctx context.Context, timesheetStatus, timesheetDate string) error {
	stepState := StepStateFromContext(ctx)
	var getTimesheetDateTime time.Time

	switch timesheetStatus {
	case "DRAFT":
		timesheetStatus = pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()
	case "SUBMITTED":
		timesheetStatus = pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()
	case "APPROVED":
		timesheetStatus = pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()
	case "CONFIRMED":
		timesheetStatus = pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()
	default:
		timesheetStatus = pb.TimesheetStatus_TIMESHEET_STATUS_NONE.String()
	}

	switch timesheetDate {
	case "YESTERDAY":
		getTimesheetDateTime = time.Now().AddDate(0, 0, -1)
	case "TOMORROW":
		getTimesheetDateTime = time.Now().AddDate(0, 0, 1)
	case "5DAYS FROM TODAY":
		getTimesheetDateTime = time.Now().AddDate(0, 0, 5)
	case "2MONTHS FROM TODAY":
		getTimesheetDateTime = time.Now().AddDate(0, 2, 0)
	default:
		getTimesheetDateTime = time.Now()
	}
	timesheetID, err := initTimesheet(ctx, stepState.CurrentUserID, locationIDs[0], timesheetStatus, getTimesheetDateTime, golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return err
	}
	stepState.CurrentTimesheetIDs = append(stepState.CurrentTimesheetIDs, timesheetID)
	stepState.CurrentTimesheetID = timesheetID
	return nil
}

func (s *Suite) createLessonRecords(ctx context.Context, lessonStatusStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	splitLessonStatus := strings.Split(lessonStatusStr, "-")
	if stepState.CurrentUserGroup != constant.UserGroupTeacher {
		err := initTeachers(ctx, stepState.CurrentUserID, stepState.CurrentSchoolID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	// create teacher record
	for _, lessonStatusFormat := range splitLessonStatus {
		var lessonStatus string
		switch lessonStatusFormat {
		case "PUBLISHED":
			lessonStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()
		case "COMPLETED":
			lessonStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()
		case "CANCELLED":
			lessonStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid lesson status")
		}
		for _, currentTimesheetID := range stepState.CurrentTimesheetIDs {
			// create lesson record
			lessonID, err := initLesson(ctx, stepState.CurrentUserID, locationIDs[0], lessonStatus, golibs.ResourcePathFromCtx(ctx))
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			// create timesheet lesson teacher record
			err = initLessonTeachers(ctx, lessonID, stepState.CurrentUserID)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			stepState.TimesheetLessonIDs = append(stepState.TimesheetLessonIDs, lessonID)
			// create timesheet lesson hours record
			err = initTimesheetLessonHours(ctx, currentTimesheetID, lessonID, golibs.ResourcePathFromCtx(ctx))
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func initTimesheetLessonHours(ctx context.Context, timesheetID, lessonID, resourcePath string) error {
	now := time.Now()
	timesheetLessonHours := new(entity.TimesheetLessonHours)
	database.AllNullEntity(timesheetLessonHours)

	if err := multierr.Combine(
		timesheetLessonHours.TimesheetID.Set(database.Text(timesheetID)),
		timesheetLessonHours.LessonID.Set(database.Text(lessonID)),
		timesheetLessonHours.UpdatedAt.Set(database.Timestamptz(now)),
		timesheetLessonHours.CreatedAt.Set(database.Timestamptz(now)),
		timesheetLessonHours.FlagOn.Set(database.Bool(false)),
	); err != nil {
		return err
	}

	fields, _ := timesheetLessonHours.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertTimesheetLessonHourFormat := fmt.Sprintf(insertTimesheetLessonHourStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(timesheetLessonHours, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertTimesheetLessonHourFormat, values...)

	if err != nil {
		return fmt.Errorf("error inserting timesheet lesson hours: %v", err)
	}

	return nil
}

func initLesson(ctx context.Context, teacherID, locationID, scheduledStatus, resourcePath string) (string, error) {
	now := time.Now()
	lesson := new(lessonmgmt_entities.Lesson)
	database.AllNullEntity(lesson)

	if err := multierr.Combine(
		lesson.LessonID.Set(database.Text(idutil.ULIDNow())),
		lesson.SchedulingStatus.Set(database.Text(scheduledStatus)),
		lesson.StreamLearnerCounter.Set(pgtype.Int4{Int: 1}),
		lesson.LearnerIds.Set([]string{idutil.ULIDNow()}),
		lesson.TeacherID.Set(teacherID),
		lesson.UpdatedAt.Set(database.Timestamptz(now)),
		lesson.CreatedAt.Set(database.Timestamptz(now)),
		lesson.IsLocked.Set(false)); err != nil {
		return "", err
	}

	fields, _ := lesson.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertLessonStatement := fmt.Sprintf(insertLessonStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(lesson, fields)

	_, err := connections.BobDBTrace.Exec(ctx, insertLessonStatement, values...)

	if err != nil {
		return "", fmt.Errorf("error inserting lesson: %v", err)
	}

	return lesson.LessonID.String, nil
}

func initLessonTeachers(ctx context.Context, lessonID, teacherID string) error {
	lessonTeacher := new(lessonmgmt_entities.LessonTeacher)
	database.AllNullEntity(lessonTeacher)
	now := time.Now()
	if err := multierr.Combine(
		lessonTeacher.LessonID.Set(database.Text(lessonID)),
		lessonTeacher.TeacherID.Set(database.Text(teacherID)),
		lessonTeacher.CreatedAt.Set(database.Timestamptz(now)),
		lessonTeacher.DeletedAt.Set(database.Timestamptz(now))); err != nil {
		return err
	}

	fields, _ := lessonTeacher.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertLessonTeacherFormat := fmt.Sprintf(insertLessonTeacherStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(lessonTeacher, fields)

	_, err := connections.BobDBTrace.Exec(ctx, insertLessonTeacherFormat, values...)

	if err != nil {
		return fmt.Errorf("error inserting lesson teacher: %v", err)
	}

	return nil
}

func (s *Suite) approvesThisTimesheet(ctx context.Context) (context.Context, error) {
	// time sleep for lesson sync before approving
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	req := &pb.ApproveTimesheetRequest{
		TimesheetIds: []string{stepState.CurrentTimesheetID},
	}

	ctx, err := s.SubscribeLockLessonEvent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("err when subscribe lock lesson event: %w", err)
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewTimesheetStateMachineServiceClient(s.TimesheetConn).ApproveTimesheet(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) SubscribeLockLessonEvent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "LockLessonSubscription",
	}
	handleLockLessonEvent := func(ctx context.Context, data []byte) (bool, error) {
		l := &pb.TimesheetLessonLockEvt{}
		err := proto.Unmarshal(data, l)
		if err != nil {
			return false, err
		}
		stepState.FoundChanForJetStream <- l
		return true, nil
	}

	sub, err := s.JSM.Subscribe(
		constants.SubjectTimesheetLesson,
		opts,
		handleLockLessonEvent,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) updateLockLessonSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	timer := time.NewTimer(time.Minute * 1)
	select {
	case data := <-stepState.FoundChanForJetStream:
		switch v := data.(type) {
		case *pb.TimesheetLessonLockEvt:
			if !stringutil.SliceElementsMatch(stepState.TimesheetLessonIDs, v.GetLessonIds()) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for lessonIds, got %s", stepState.TimesheetLessonIDs, v.GetLessonIds())
			}
			if len(v.GetLessonIds()) > 0 {
				e := &lessonmgmt_entities.LessonRepo{}
				for _, l := range v.GetLessonIds() {
					var lesson *domain.Lesson
					err := try.Do(func(attempt int) (retry bool, err error) {
						lesson, err = e.GetLessonByID(ctx, s.BobDB, l)
						if err != nil {
							return false, err
						}
						if lesson.IsLocked {
							return false, nil
						} else {
							time.Sleep(10 * time.Second)
							return attempt < 10, fmt.Errorf("need IsLocked from before scenario")
						}
					})
					if err != nil {
						return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s have isLocked %t: %w", lesson.LessonID, lesson.IsLocked, err)
					}
				}
				return StepStateToContext(ctx, stepState), nil
			}
			return StepStateToContext(ctx, stepState), fmt.Errorf("v.GetLessonIDs from message is empty")
		}

	case <-ctx.Done():
		return StepStateToContext(ctx, stepState), fmt.Errorf("timeout waiting for event to be published")
	case <-timer.C:
		return StepStateToContext(ctx, stepState), errors.New("time out cause of failing")
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("error updateLockLessonSuccessfully")
}

func (s *Suite) timesheetStatusChangedToApprove(ctx context.Context, approvedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch approvedStatus {
	case "successfully":
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if stepState.Response != nil {
			if !stepState.Response.(*pb.ApproveTimesheetResponse).Success {
				return ctx, fmt.Errorf("error cannot approved timesheet record")
			}

			err := s.checkTimesheetRecord(ctx, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	case "unsuccessfully":
		if stepState.ResponseErr == nil {
			return ctx, fmt.Errorf("error timesheet record should not be approved")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkTimesheetRecord(ctx context.Context, timesheetStatus string) error {
	stepState := StepStateFromContext(ctx)
	var count int
	stmt := `
		SELECT
			count(timesheet_id)
		FROM
			timesheet
		WHERE
			timesheet_id = $1
		AND
			deleted_at IS NULL
		AND
			timesheet_status = $2
		`
	err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentTimesheetID, timesheetStatus).Scan(&count)
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("unexpected %d timesheet record", count)
	}

	return nil
}

func initStaff(ctx context.Context, userId string, resourcePath string) error {
	staff := &entity.Staff{
		StaffID:   database.Text(userId),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}

	fields, _ := staff.FieldMap()
	fields = append(fields, "resource_path")
	insertStaffStatement := fmt.Sprintf(insertStaffStmtFormat, strings.Join(fields, ","), database.GeneratePlaceholders(len(fields)))

	values := database.GetScanFields(staff, fields)
	values = append(values, database.Text(resourcePath))
	_, err := connections.TimesheetDB.Exec(ctx, insertStaffStatement, values...)
	if err != nil {
		return fmt.Errorf("error inserting staff, userID: %v,error: %v", userId, err)
	}
	return nil
}

func initTimesheet(ctx context.Context, userId string, locationID string, status string, timesheetDate time.Time, resourcePath string) (string, error) {
	timesheet := entity.NewTimesheet()
	timesheet.StaffID = database.Text(userId)
	timesheet.TimesheetDate = database.Timestamptz(timesheetDate)
	timesheet.UpdatedAt = database.Timestamptz(time.Now())
	timesheet.CreatedAt = database.Timestamptz(time.Now())
	timesheet.LocationID = database.Text(locationID)
	timesheet.TimesheetStatus = database.Text(status)

	fields, _ := timesheet.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertStaffStatement := fmt.Sprintf(insertTimesheetStmtFormat, strings.Join(fields, ","), placeHolder)

	values := database.GetScanFields(timesheet, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertStaffStatement, values...)

	if err != nil {
		return "", fmt.Errorf("error inserting timesheet, userID: %v,error: %v", userId, err)
	}

	return timesheet.TimesheetID.String, err
}

func initTeachers(ctx context.Context, teacherID string, resourcePath int32) error {
	teacher := new(user_entities.Teacher)
	database.AllNullEntity(teacher)

	teacher.ID.Set(database.Text(teacherID))
	now := time.Now()
	if err := teacher.UpdatedAt.Set(now); err != nil {
		return err
	}
	if err := teacher.CreatedAt.Set(now); err != nil {
		return err
	}

	fields, _ := teacher.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertStaffFormat := fmt.Sprintf(insertStaffStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(teacher, fields)

	_, err := connections.BobDBTrace.Exec(ctx, insertStaffFormat, values...)

	if err != nil {
		return fmt.Errorf("error inserting staff: %v", err)
	}
	// will remove after remove table teacher
	t := &entities.Teacher{}
	database.AllNullEntity(t)
	t.ID.Set(database.Text(teacherID))
	t.UpdatedAt = teacher.UpdatedAt
	t.CreatedAt = teacher.CreatedAt
	fields, _ = t.FieldMap()
	placeHolder = database.GeneratePlaceholders(len(fields))
	insertTeacherFormat := fmt.Sprintf(insertTeacherStmtFormat, strings.Join(fields, ","), placeHolder)
	values = database.GetScanFields(t, fields)
	_, err = connections.BobDBTrace.Exec(ctx, insertTeacherFormat, values...)
	if err != nil {
		return fmt.Errorf("error inserting teacher: %v", err)
	}

	if err != nil {
		return fmt.Errorf("error inserting teacher: %v", err)
	}

	return nil
}

func initLocation(ctx context.Context, locationIDs []string) error {
	for _, locationID := range locationIDs {
		stmt := `INSERT INTO locations (location_id,name, is_archived) VALUES($1,$2,$3)
				ON CONFLICT DO NOTHING`
		_, err := connections.TimesheetDB.Exec(ctx, stmt, locationID, locationID,
			false)
		if err != nil {
			return err
		}
	}
	return nil
}
