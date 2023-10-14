package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	nats_service_utils "github.com/manabie-com/backend/internal/timesheet/service/nats"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimesheetStateMachineService struct {
	DB            database.Ext
	JSM           nats.JetStreamManagement
	TimesheetRepo interface {
		FindTimesheetByTimesheetID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Timesheet, error)
		SoftDeleteByIDs(ctx context.Context, db database.QueryExecer, id pgtype.TextArray) error
		UpdateTimeSheet(ctx context.Context, db database.QueryExecer, timesheet *entity.Timesheet) (*entity.Timesheet, error)
		FindTimesheetByTimesheetIDsAndStatus(ctx context.Context, db database.QueryExecer, ids []string, timesheetStatus string) ([]*entity.Timesheet, error)
		UpdateTimesheetStatusMultiple(ctx context.Context, db database.QueryExecer, timesheets []*entity.Timesheet, timesheetStatus string) error
	}
	TimesheetLessonHoursRepo interface {
		FindTimesheetLessonHoursByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) ([]*entity.TimesheetLessonHours, error)
		FindByTimesheetIDs(ctx context.Context, db database.QueryExecer, timesheetIDs []string) ([]*entity.TimesheetLessonHours, error)
	}
	OtherWorkingHoursRepo interface {
		SoftDeleteByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) error
	}
	LessonRepo interface {
		FindLessonsByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.Lesson, error)
	}
	TransportationExpenseRepo interface {
		SoftDeleteByTimesheetID(ctx context.Context, db database.QueryExecer, timesheetID pgtype.Text) error
	}
}

func (s *TimesheetStateMachineService) DeleteTimesheet(ctx context.Context, timesheetID string) error {
	timesheet, err := s.TimesheetRepo.FindTimesheetByTimesheetID(ctx, s.DB, database.Text(timesheetID))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", err.Error()))
	}

	if timesheet.TimesheetStatus.String == pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String() {
		err = s.validateTimesheetPermission(ctx, timesheet)
		if err != nil {
			return err
		}
	} else if timesheet.TimesheetStatus.String == pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String() {
		err := checkPermissionForApproverConfirmer(ctx)
		if err != nil {
			return err
		}
	} else {
		return status.Error(codes.FailedPrecondition, fmt.Sprintf("can not delete timesheet %s in : %s status", timesheet.TimesheetID.String, timesheet.TimesheetStatus.String))
	}

	// check for timesheet lesson hours
	timesheetLessonHours, err := s.TimesheetLessonHoursRepo.FindTimesheetLessonHoursByTimesheetID(ctx, s.DB, database.Text(timesheetID))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find timesheet lesson hours error: %s", err.Error()))
	}

	if len(timesheetLessonHours) > 0 {
		for _, elm := range timesheetLessonHours {
			if elm.FlagOn.Bool {
				return status.Error(codes.FailedPrecondition, "timesheet record has lesson record")
			}
		}
	}

	// Check timesheet is the past
	isThePastTimesheet := false
	timeNow := time.Now().In(timeutil.Timezone(pbc.COUNTRY_JP)) // date in Japan timezone
	dateNow := time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), 0, 0, 0, 0, timeNow.Location())

	if timesheet.TimesheetDate.Time.In(timeutil.Timezone(pbc.COUNTRY_JP)).Before(dateNow) {
		isThePastTimesheet = true
	}
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// Delete Other Working hours
		if err := s.OtherWorkingHoursRepo.SoftDeleteByTimesheetID(ctx, tx, database.Text(timesheetID)); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("delete other working hours error: %s", err.Error()))
		}

		// Delete Transport expenses
		if err := s.TransportationExpenseRepo.SoftDeleteByTimesheetID(ctx, tx, database.Text(timesheetID)); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("delete transport expenses error: %s", err.Error()))
		}

		// Delete the timesheet
		if isThePastTimesheet || len(timesheetLessonHours) == 0 {
			if err := s.TimesheetRepo.SoftDeleteByIDs(ctx, tx, database.TextArray([]string{timesheetID})); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("delete timesheet error: %s", err.Error()))
			}
		}

		return nil
	})

	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *TimesheetStateMachineService) SubmitTimesheet(ctx context.Context, timesheetID string) error {
	timesheet, err := s.TimesheetRepo.FindTimesheetByTimesheetID(ctx, s.DB, database.Text(timesheetID))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", err.Error()))
	}

	err = s.validateTimesheetPermissionAndStatus(ctx, timesheet)
	if err != nil {
		return err
	}

	err = validateSubmitTimesheetDate(timesheet.TimesheetDate.Time)
	if err != nil {
		return err
	}

	timesheetLessonHours, err := s.TimesheetLessonHoursRepo.FindTimesheetLessonHoursByTimesheetID(ctx, s.DB, database.Text(timesheetID))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find timesheet lesson hours error: %s", err.Error()))
	}

	// validate lesson status if there's timesheet lesson records
	if len(timesheetLessonHours) > 0 {
		var lessonIDs []string
		for _, timesheetLesson := range timesheetLessonHours {
			// only get lesson hour with flag on = true
			if timesheetLesson.FlagOn.Bool {
				lessonIDs = append(lessonIDs, timesheetLesson.LessonID.String)
			}
		}
		lessons, err := s.LessonRepo.FindLessonsByIDs(ctx, s.DB, database.TextArray(lessonIDs))
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("find lesson record error: %s", err.Error()))
		}
		// invalid lesson status if not completed or cancelled
		err = validateLessonStatus(lessons)
		if err != nil {
			return err
		}
	}

	// update timesheet status to submitted
	err = s.updateSubmittedTimesheetStatus(ctx, timesheet)
	if err != nil {
		return err
	}

	timeExecuted := time.Now()
	// send submit timesheet event to NATS
	msg := &pb.TimesheetActionLogRequest{
		Action:      pb.TimesheetAction_SUBMITTED,
		ExecutedBy:  interceptors.UserIDFromContext(ctx),
		TimesheetId: timesheetID,
		IsSystem:    false,
		ExecutedAt:  timestamppb.New(timeExecuted),
	}
	err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *TimesheetStateMachineService) ApproveTimesheet(ctx context.Context, timesheetIDs []string) error {
	err := checkPermissionForApproverConfirmer(ctx)
	if err != nil {
		return err
	}
	// find submitted timesheets by timesheet ids
	timesheets, err := s.TimesheetRepo.FindTimesheetByTimesheetIDsAndStatus(ctx, s.DB, timesheetIDs, pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String())
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find submitted timesheet error: %s", err.Error()))
	}

	// length of find submitted timesheet is not equal to the request timesheet
	if len(timesheets) != len(timesheetIDs) {
		return status.Error(codes.Internal, "find submitted timesheet records not match with the request")
	}

	// from timesheet => find timesheet lesson hour => lesson => validate lesson status
	for _, timesheet := range timesheets {
		timesheetLessonHours, err := s.TimesheetLessonHoursRepo.FindTimesheetLessonHoursByTimesheetID(ctx, s.DB, database.Text(timesheet.TimesheetID.String))
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("find timesheet lesson hours error: %s", err.Error()))
		}

		// validate lesson status if there's timesheet lesson records
		if len(timesheetLessonHours) > 0 {
			var lessonIDs []string
			for _, timesheetLesson := range timesheetLessonHours {
				// only get lesson hour with flag on = true
				if timesheetLesson.FlagOn.Bool {
					lessonIDs = append(lessonIDs, timesheetLesson.LessonID.String)
				}
			}

			lessons, err := s.LessonRepo.FindLessonsByIDs(ctx, s.DB, database.TextArray(lessonIDs))
			if err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("find lesson record error: %s", err.Error()))
			}
			// invalid lesson status if not completed or cancelled
			err = validateLessonStatusToApproveTimesheet(lessons)
			if err != nil {
				return err
			}
		}
	}

	// set timesheets approve status
	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		err = s.TimesheetRepo.UpdateTimesheetStatusMultiple(ctx, tx, timesheets, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("upsert multiple timesheet error: %s", err.Error()))
		}
		// find lesson ids with timesheet ids
		timesheetLessonHours, err := s.TimesheetLessonHoursRepo.FindByTimesheetIDs(ctx, tx, timesheetIDs)
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("find by timesheet ids lesson hours error: %s", err.Error()))
		}

		var lessonIDs []string
		for _, timesheetLessonHour := range timesheetLessonHours {
			lessonIDs = append(lessonIDs, timesheetLessonHour.LessonID.String)
		}

		// send bulk lesson lock flag event
		msg := &pb.TimesheetLessonLockEvt{
			LessonIds: lessonIDs,
		}
		err = s.PublishLockLessonEvent(ctx, msg)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		return nil
	})

	if err != nil {
		return err
	}

	timeExecuted := time.Now()
	// send approve timesheet event to NATS
	for _, timesheetID := range timesheetIDs {
		msg := &pb.TimesheetActionLogRequest{
			Action:      pb.TimesheetAction_APPROVED,
			ExecutedBy:  interceptors.UserIDFromContext(ctx),
			TimesheetId: timesheetID,
			IsSystem:    false,
			ExecutedAt:  timestamppb.New(timeExecuted),
		}
		err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (s *TimesheetStateMachineService) PublishLockLessonEvent(ctx context.Context, msg *pb.TimesheetLessonLockEvt) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgID, err := s.JSM.PublishAsyncContext(ctx, constants.SubjectTimesheetLesson, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishLockLessonEvent JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}

func (s *TimesheetStateMachineService) CancelApproveTimesheet(ctx context.Context, timesheetID string) error {
	err := checkPermissionForApproverConfirmer(ctx)
	if err != nil {
		return err
	}

	timesheet, err := s.TimesheetRepo.FindTimesheetByTimesheetID(ctx, s.DB, database.Text(timesheetID))
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", err.Error()))
	}

	if timesheet.TimesheetStatus.String != pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String() {
		return status.Error(codes.FailedPrecondition, "timesheet status should be in Approved")
	}

	err = s.updateSubmittedTimesheetStatus(ctx, timesheet)
	if err != nil {
		return err
	}

	timeExecuted := time.Now()
	// send cancel approve timesheet action log event to NATS
	msg := &pb.TimesheetActionLogRequest{
		Action:      pb.TimesheetAction_CANCEL_APPROVAL,
		ExecutedBy:  interceptors.UserIDFromContext(ctx),
		TimesheetId: timesheetID,
		IsSystem:    false,
		ExecutedAt:  timestamppb.New(timeExecuted),
	}
	err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *TimesheetStateMachineService) ConfirmTimesheet(ctx context.Context, timesheetIDs []string) error {
	err := checkPermissionForApproverConfirmer(ctx)
	if err != nil {
		return err
	}

	// find approved timesheets by timesheet ids
	timesheets, err := s.TimesheetRepo.FindTimesheetByTimesheetIDsAndStatus(ctx, s.DB, timesheetIDs, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find approved timesheet error: %s", err.Error()))
	}

	// length of find submitted timesheet is not equal to the request timesheet
	if len(timesheets) != len(timesheetIDs) {
		return status.Error(codes.Internal, "find approved timesheet records not match with the request")
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		// set timesheets confirm status
		err = s.TimesheetRepo.UpdateTimesheetStatusMultiple(ctx, tx, timesheets, pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String())
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("upsert multiple timesheet status to confirm error: %s", err.Error()))
		}
		return nil
	})

	if err != nil {
		return err
	}

	timeExecuted := time.Now()
	// send confirm timesheet event to NATS
	for _, timesheetID := range timesheetIDs {
		msg := &pb.TimesheetActionLogRequest{
			Action:      pb.TimesheetAction_CONFIRMED,
			ExecutedBy:  interceptors.UserIDFromContext(ctx),
			TimesheetId: timesheetID,
			IsSystem:    false,
			ExecutedAt:  timestamppb.New(timeExecuted),
		}
		err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (s *TimesheetStateMachineService) updateSubmittedTimesheetStatus(ctx context.Context, timesheet *entity.Timesheet) error {
	timesheet.TimesheetStatus.Set(database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()))
	if _, err := s.TimesheetRepo.UpdateTimeSheet(ctx, s.DB, timesheet); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("update timesheet status submitted error: %s", err.Error()))
	}

	return nil
}

func (s *TimesheetStateMachineService) validateTimesheetPermissionAndStatus(ctx context.Context, timesheet *entity.Timesheet) error {
	if err := checkPermissionToModifyTimesheet(ctx, timesheet.StaffID.String); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	if timesheet.TimesheetStatus.String != pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String() {
		return status.Error(codes.FailedPrecondition, "timesheet status should be in Draft")
	}

	return nil
}

func validateSubmitTimesheetDate(timesheetDate time.Time) error {
	dateFormat := "2006-01-02 00:00:00"
	// timesheet date for future is invalid
	if timesheetDate.Format(dateFormat) > time.Now().Format(dateFormat) {
		return status.Error(codes.FailedPrecondition, "timesheet date should not be in future")
	}

	return nil
}

func validateLessonStatus(lessons []*entity.Lesson) error {
	for _, lesson := range lessons {
		if lesson.SchedulingStatus.String != cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String() && lesson.SchedulingStatus.String != cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String() {
			return status.Error(codes.FailedPrecondition, "invalid lesson status")
		}
	}

	return nil
}

func checkPermissionForApproverConfirmer(ctx context.Context) error {
	userRoles := interceptors.UserRolesFromContext(ctx)
	for _, role := range userRoles {
		if _, found := constant.RolesWriteOtherMemberTimesheet[role]; found {
			return nil
		}
	}

	return status.Error(codes.PermissionDenied, fmt.Sprintf("unauthorized to modify timesheet, timesheetStaffID: %s", interceptors.UserIDFromContext(ctx)))
}

func (s *TimesheetStateMachineService) CancelSubmissionTimesheet(ctx context.Context, timesheetID string) error {
	timesheet, err := s.TimesheetRepo.FindTimesheetByTimesheetID(ctx, s.DB, database.Text(timesheetID))

	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("find timesheet error: %s", err.Error()))
	}

	err = s.validateTimesheetPermission(ctx, timesheet)
	if err != nil {
		return err
	}

	if timesheet.TimesheetStatus.String != pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String() {
		return status.Error(codes.FailedPrecondition, "timesheet status should be in Submitted")
	}

	err = s.updateDraftTimesheetStatus(ctx, timesheet)
	if err != nil {
		return err
	}

	timeExecuted := time.Now()
	// send cancel submission action log event to NATS
	msg := &pb.TimesheetActionLogRequest{
		Action:      pb.TimesheetAction_CANCEL_SUBMISSION,
		ExecutedBy:  interceptors.UserIDFromContext(ctx),
		TimesheetId: timesheetID,
		IsSystem:    false,
		ExecutedAt:  timestamppb.New(timeExecuted),
	}
	err = nats_service_utils.PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	return nil
}

func (s *TimesheetStateMachineService) updateDraftTimesheetStatus(ctx context.Context, timesheet *entity.Timesheet) error {
	timesheet.TimesheetStatus.Set(database.Text(pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()))
	if _, err := s.TimesheetRepo.UpdateTimeSheet(ctx, s.DB, timesheet); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("update timesheet status draft error: %s", err.Error()))
	}

	return nil
}

func (s *TimesheetStateMachineService) validateTimesheetPermission(ctx context.Context, timesheet *entity.Timesheet) error {
	if err := checkPermissionToModifyTimesheet(ctx, timesheet.StaffID.String); err != nil {
		return status.Error(codes.PermissionDenied, err.Error())
	}

	return nil
}

func validateLessonStatusToApproveTimesheet(lessons []*entity.Lesson) error {
	for _, lesson := range lessons {
		if !(lesson.SchedulingStatus.String == cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String() || lesson.SchedulingStatus.String == cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()) {
			return status.Error(codes.FailedPrecondition, "invalid lesson status")
		}
	}

	return nil
}
