package nats

import (
	"context"
	"errors"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/timesheet/domain/common"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	v11 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LessonNatsServiceImpl struct {
	DB  database.Ext
	JSM nats.JetStreamManagement

	GetTimesheetService interface {
		GetTimesheet(ctx context.Context, timesheetQueryArgs *dto.TimesheetQueryArgs, timesheetQueryOptions *dto.TimesheetGetOptions) ([]*dto.Timesheet, error)
		GetTimesheetByLessonIDs(ctx context.Context, lessonIds []string) ([]*dto.Timesheet, error)
	}
	AutoCreateTimesheetService interface {
		CreateTimesheetMultiple(ctx context.Context, timesheets []*dto.Timesheet) ([]*dto.Timesheet, error)
		RemoveTimesheetLessonHoursMultiple(ctx context.Context, timesheets []*dto.Timesheet) ([]*dto.Timesheet, error)
		CreateAndRemoveTimesheetMultiple(ctx context.Context, newTimesheets, removeTimesheets []*dto.Timesheet) error
		UpdateLessonAutoCreateFlagState(ctx context.Context, flagsMap map[bool][]string) error
	}
	AutoCreateFlagActivityLogService interface {
		GetAutoCreateFlagLogByTeacherIDs(ctx context.Context, lessonStartDate time.Time, teacherIDs []string) ([]*dto.AutoCreateFlagActivityLog, error)
	}
	ConfirmationWindowService interface {
		CheckModifyConditionByTimesheetDateAndLocation(ctx context.Context, timesheetDate *timestamppb.Timestamp, locationID string) (bool, error)
		CheckModifyConditionByTimesheetID(ctx context.Context, timesheetID string) (bool, error)
	}
	LessonRepo interface {
		FindAllLessonsByIDsIgnoreDeletedAtCondition(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entity.Lesson, error)
	}
	PartnerAutoCreateTimesheetRepo interface {
		GetPartnerAutoCreateDefaultValue(ctx context.Context, db database.QueryExecer) (*entity.PartnerAutoCreateTimesheetFlag, error)
	}
}

func (s *LessonNatsServiceImpl) HandleEventLessonDelete(ctx context.Context, msg *bpb.EvtLesson_DeletedLessons) error {
	var (
		zapLogger         = ctxzap.Extract(ctx).Sugar()
		lessonIDsToDelete []string
	)
	lessonIDs := msg.LessonIds
	if len(lessonIDs) < 1 {
		zapLogger.Warn("HandleEventLessonDelete DeletedLessonIDs is empty")
		return nil
	}
	lessons, err := s.LessonRepo.FindAllLessonsByIDsIgnoreDeletedAtCondition(ctx, s.DB, database.TextArray(lessonIDs))
	if err != nil {
		zapLogger.Errorf("HandleEventLessonDelete FindAllLessonsByIDsIgnoreDeletedAtCondition failed, err: %v, lessonIDs: %v", err, msg.LessonIds)
		return err
	}

	for _, lesson := range lessons {
		isAllowToModify, err := s.ConfirmationWindowService.CheckModifyConditionByTimesheetDateAndLocation(ctx, &timestamppb.Timestamp{Seconds: lesson.StartTime.Time.Unix()}, lesson.CenterID.String)
		if err != nil {
			zapLogger.Errorf("HandleEventLessonDelete CheckModifyConditionByTimesheetDateAndLocation failed, err: %v", err)
			return err
		}
		if !isAllowToModify {
			continue
		}
		lessonIDsToDelete = append(lessonIDsToDelete, lesson.LessonID.String)
	}
	if len(lessonIDsToDelete) < 1 {
		zapLogger.Warn("HandleEventLessonDelete DeletedLessonIDs after check modify condition is empty")
		return nil
	}

	timesheets, err := s.GetTimesheetService.GetTimesheetByLessonIDs(ctx, lessonIDsToDelete)
	if err != nil {
		zapLogger.Errorf("HandleEventLessonDelete GetTimesheetByLessonIDs failed, err: %v, lessonIDs: %v", err, lessonIDsToDelete)
		return err
	}
	for _, lessonID := range lessonIDsToDelete {
		timesheets = removeTimesheetLessonHours(ctx, lessonID, timesheets)
	}
	_, err = s.AutoCreateTimesheetService.RemoveTimesheetLessonHoursMultiple(ctx, timesheets)
	if err != nil {
		zapLogger.Errorf("HandleEventLessonDelete RemoveTimesheetLessonHoursMultiple failed, err: %v", err)
		return err
	}

	if err = s.publishSystemUpdateLessonActionLogEvent(ctx, timesheets); err != nil {
		zapLogger.Errorf("HandleEventLessonDelete publishSystemUpdateLessonActionLogEvent failed, err: %v", err)
		return err
	}
	return nil
}

func (s *LessonNatsServiceImpl) HandleEventLessonUpdate(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error {
	var (
		updateTypes = checkLessonUpdateType(msg)
		zapLogger   = ctxzap.Extract(ctx).Sugar()
	)

	for _, updateType := range updateTypes {
		isAllowToModifyBeforeUpdate, err := s.ConfirmationWindowService.CheckModifyConditionByTimesheetDateAndLocation(ctx, &timestamppb.Timestamp{Seconds: msg.StartAtBefore.AsTime().Unix()}, msg.LocationIdBefore)
		if err != nil {
			zapLogger.Errorf("HandleEventLessonUpdate CheckModifyConditionByTimesheetDateAndLocation Before Update failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
				err, msg.TeacherIdsBefore, msg.LocationIdBefore, msg.StartAtBefore.AsTime())
			return err
		}
		isAllowToModifyAfterUpdate, err := s.ConfirmationWindowService.CheckModifyConditionByTimesheetDateAndLocation(ctx, &timestamppb.Timestamp{Seconds: msg.StartAtAfter.AsTime().Unix()}, msg.LocationIdAfter)
		if err != nil {
			zapLogger.Errorf("HandleEventLessonUpdate CheckModifyConditionByTimesheetDateAndLocation After Update failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
				err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
			return err
		}
		if !isAllowToModifyBeforeUpdate && !isAllowToModifyAfterUpdate {
			continue
		}
		switch updateType {
		case constant.LessonUpdateTypeDraftToPublished:
			timesheets, err := s.GetTimesheetService.GetTimesheet(ctx,
				&dto.TimesheetQueryArgs{
					StaffIDs:      msg.TeacherIdsAfter,
					LocationID:    msg.LocationIdAfter,
					TimesheetDate: msg.StartAtAfter.AsTime(),
				},
				&dto.TimesheetGetOptions{
					IsGetListTimesheetLessonHours: true,
				},
			)

			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate DraffToPublish GetTimesheet failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
					err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
				return err
			}

			mapAutoCreateFlagLog, err := s.mappingStaffAndAutoCreateFlagActivityLog(ctx, msg.StartAtAfter.AsTime(), msg.TeacherIdsAfter)

			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate DraffToPublish mappingStaffAndAutoCreateFlagActivityLog failed, err: %v, staffIDs: %v, timesheetDate: %v", err, msg.TeacherIdsAfter, msg.StartAtAfter.AsTime())
				return err
			}

			partnerAutoCreateDefaultValue := constant.KPartnerAutoCreateDefaultValue
			partnerAutoCreateDefaultE, err := s.PartnerAutoCreateTimesheetRepo.GetPartnerAutoCreateDefaultValue(ctx, s.DB)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				zapLogger.Errorf("HandleEventLessonCreate GetPartnerAutoCreateDefaultValue failed, err: %v", err)
				return err
			}

			if partnerAutoCreateDefaultE != nil {
				partnerAutoCreateDefaultValue = partnerAutoCreateDefaultE.FlagOn.Bool
			}

			timesheetDatas := buildTimesheetData(
				msg.LessonId,
				msg.LocationIdAfter,
				msg.TeacherIdsAfter,
				msg.StartAtAfter.AsTime(),
				timesheets,
				mapAutoCreateFlagLog,
				partnerAutoCreateDefaultValue,
			)

			_, err = s.AutoCreateTimesheetService.CreateTimesheetMultiple(ctx, timesheetDatas)
			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate DraffToPublish CreateTimesheetMultiple failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
					err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
				return err
			}

			if err = s.publishSystemUpdateLessonActionLogEvent(ctx, timesheetDatas); err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate DraffToPublish publishSystemUpdateLessonActionLogEvent failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
					err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
				return err
			}
		case constant.LessonUpdateTypePublishedToDraft:
			timesheets, err := s.GetTimesheetService.GetTimesheet(ctx,
				&dto.TimesheetQueryArgs{
					StaffIDs:      msg.TeacherIdsBefore,
					LocationID:    msg.LocationIdBefore,
					TimesheetDate: msg.StartAtBefore.AsTime(),
				},
				&dto.TimesheetGetOptions{
					IsGetListOtherWorkingHours:     true,
					IsGetListTimesheetLessonHours:  true,
					IsGetListTransportationExpense: true,
				},
			)

			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate PublishedToDraft GetTimesheet, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
					err, msg.TeacherIdsBefore, msg.LocationIdBefore, msg.StartAtBefore.AsTime())
				return err
			}

			timesheets = removeTimesheetLessonHours(ctx, msg.LessonId, timesheets)

			_, err = s.AutoCreateTimesheetService.RemoveTimesheetLessonHoursMultiple(ctx, timesheets)
			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate PublishedToDraft RemoveTimesheetLessonHoursMultiple failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
					err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
				return err
			}

			if err = s.publishSystemUpdateLessonActionLogEvent(ctx, timesheets); err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate PublishedToDraft publishSystemUpdateLessonActionLogEvent failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
					err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
				return err
			}
		case constant.LessonUpdateTypeChangeLessonStartTime:
			err = s.lessonUpdateChangeLessonStartTime(ctx, msg)
			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate ChangeStartTime lessonUpdateChangeStartTime failed, err: %v", err)
				return err
			}
		case constant.LessonUpdateTypeChangeTeacherID:
			err = s.lessonUpdateChangeTeachers(ctx, msg)
			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate ChangeTeacher lessonUpdateChangeTeachers failed, err: %v", err)
				return err
			}
		case constant.LessonUpdateTypeChangeLessonDate:
			err = s.lessonUpdateChangeLessonDate(ctx, msg)
			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate ChangeLessonDate lessonUpdateChangeLessonDate failed, err: %v", err)
				return err
			}
		case constant.LessonUpdateTypeChangeLocationID:
			err = s.lessonUpdateChangeLocation(ctx, msg)
			if err != nil {
				zapLogger.Errorf("HandleEventLessonUpdate ChangeLocationID lessonUpdateChangeLocation failed, err: %v", err)
				return err
			}
		default:
			zapLogger.Warnf("HandleEventLessonUpdate: update lesson type not valid: %v", updateType)
		}
	}
	return nil
}

func (s *LessonNatsServiceImpl) HandleEventLessonCreate(ctx context.Context, msg *bpb.EvtLesson_CreateLessons) error {
	var (
		zapLogger                     = ctxzap.Extract(ctx).Sugar()
		timesheetModified             []*dto.Timesheet
		partnerAutoCreateDefaultValue = constant.KPartnerAutoCreateDefaultValue
	)
	for _, lesson := range msg.Lessons {
		if lesson.SchedulingStatus != v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED {
			continue
		}

		if len(lesson.TeacherIds) == 0 {
			zapLogger.Warn("HandleEventLessonCreate lesson not contains any teachers, lessonID: %v", lesson.LessonId)
			continue
		}

		isAllowToModify, err := s.ConfirmationWindowService.CheckModifyConditionByTimesheetDateAndLocation(ctx, lesson.StartAt, lesson.LocationId)
		if err != nil {
			zapLogger.Errorf("HandleEventLessonCreate CheckModifyConditionByTimesheetDateAndLocation failed, err: %v", err)
			return err
		}
		if !isAllowToModify {
			continue
		}

		timesheets, err := s.GetTimesheetService.GetTimesheet(ctx,
			&dto.TimesheetQueryArgs{
				StaffIDs:      lesson.TeacherIds,
				LocationID:    lesson.LocationId,
				TimesheetDate: lesson.StartAt.AsTime(),
			},
			&dto.TimesheetGetOptions{
				IsGetListTimesheetLessonHours: true,
			},
		)
		if err != nil {
			zapLogger.Errorf("HandleEventLessonCreate GetTimesheet failed, err: %v", err)
			return err
		}

		mapAutoCreateFlagLog, err := s.mappingStaffAndAutoCreateFlagActivityLog(ctx, lesson.StartAt.AsTime(), lesson.TeacherIds)

		if err != nil {
			zapLogger.Errorf("HandleEventLessonCreate mappingStaffAndAutoCreateFlagActivityLog failed, err: %v", err)
			return err
		}

		partnerAutoCreateDefaultE, err := s.PartnerAutoCreateTimesheetRepo.GetPartnerAutoCreateDefaultValue(ctx, s.DB)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			zapLogger.Errorf("HandleEventLessonCreate GetPartnerAutoCreateDefaultValue failed, err: %v", err)
			return err
		}

		if partnerAutoCreateDefaultE != nil {
			partnerAutoCreateDefaultValue = partnerAutoCreateDefaultE.FlagOn.Bool
		}

		timesheetModified, err = common.MergeListTimesheet(timesheetModified, buildTimesheetDataFromLesson(lesson, timesheets, mapAutoCreateFlagLog, partnerAutoCreateDefaultValue))
		if err != nil {
			zapLogger.Errorf("HandleEventLessonCreate MergeListTimesheet failed, err: %v", err)
			return err
		}
	}

	_, err := s.AutoCreateTimesheetService.CreateTimesheetMultiple(ctx, timesheetModified)
	if err != nil {
		zapLogger.Errorf("HandleEventLessonCreate CreateTimesheetMultiple failed, err: %v", err)
		return err
	}

	if err = s.publishSystemUpdateLessonActionLogEvent(ctx, timesheetModified); err != nil {
		zapLogger.Errorf("HandleEventLessonCreate publishSystemUpdateLessonActionLogEvent failed, err: %v", err)
		return err
	}

	return nil
}

func (s *LessonNatsServiceImpl) lessonUpdateChangeTeachers(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error {
	// calculate added and removed teachers
	newTeachers, removedTeachers := calculateLessonTeachersChanged(msg.TeacherIdsBefore, msg.TeacherIdsAfter)

	return s.createAndRemoveTimesheetWhenLessonChanged(
		ctx,
		msg.GetLessonId(),
		removedTeachers, newTeachers,
		msg.GetLocationIdBefore(), msg.GetLocationIdAfter(),
		msg.GetStartAtBefore().AsTime(), msg.GetStartAtAfter().AsTime(),
	)
}

func (s *LessonNatsServiceImpl) lessonUpdateChangeLessonDate(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error {
	return s.createAndRemoveTimesheetWhenLessonChanged(
		ctx,
		msg.GetLessonId(),
		msg.GetTeacherIdsBefore(), msg.GetTeacherIdsAfter(),
		msg.GetLocationIdBefore(), msg.GetLocationIdAfter(),
		msg.GetStartAtBefore().AsTime(), msg.GetStartAtAfter().AsTime(),
	)
}

func (s *LessonNatsServiceImpl) lessonUpdateChangeLocation(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error {
	return s.createAndRemoveTimesheetWhenLessonChanged(
		ctx,
		msg.GetLessonId(),
		msg.GetTeacherIdsBefore(), msg.GetTeacherIdsAfter(),
		msg.GetLocationIdBefore(), msg.GetLocationIdAfter(),
		msg.GetStartAtBefore().AsTime(), msg.GetStartAtAfter().AsTime(),
	)
}

func (s *LessonNatsServiceImpl) lessonUpdateChangeLessonStartTime(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error {
	var (
		zapLogger                     = ctxzap.Extract(ctx).Sugar()
		partnerAutoCreateDefaultValue = constant.KPartnerAutoCreateDefaultValue
	)

	timesheets, err := s.GetTimesheetService.GetTimesheet(ctx,
		&dto.TimesheetQueryArgs{
			StaffIDs:      msg.TeacherIdsAfter,
			LocationID:    msg.LocationIdAfter,
			TimesheetDate: msg.StartAtAfter.AsTime(),
		},
		&dto.TimesheetGetOptions{},
	)

	if err != nil {
		zapLogger.Errorf("HandleEventLessonUpdate lessonUpdateChangeLessonStartTime GetTimesheet failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
			err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
		return err
	}

	mapAutoCreateFlagLog, err := s.mappingStaffAndAutoCreateFlagActivityLog(ctx, msg.StartAtAfter.AsTime(), msg.TeacherIdsAfter)

	if err != nil {
		zapLogger.Errorf("HandleEventLessonUpdate lessonUpdateChangeLessonStartTime mappingStaffAndAutoCreateFlagActivityLog failed, err: %v, staffIDs: %v, timesheetDate: %v", err, msg.TeacherIdsAfter, msg.StartAtAfter.AsTime())
		return err
	}

	partnerAutoCreateDefaultE, err := s.PartnerAutoCreateTimesheetRepo.GetPartnerAutoCreateDefaultValue(ctx, s.DB)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		zapLogger.Errorf("HandleEventLessonCreate GetPartnerAutoCreateDefaultValue failed, err: %v", err)
		return err
	}

	if partnerAutoCreateDefaultE != nil {
		partnerAutoCreateDefaultValue = partnerAutoCreateDefaultE.FlagOn.Bool
	}

	// get map[flagon]=[]string{stafIDs}
	flagsMap := buildTimesheetLessonAutoCreateFlagMap(timesheets, mapAutoCreateFlagLog, partnerAutoCreateDefaultValue)

	err = s.AutoCreateTimesheetService.UpdateLessonAutoCreateFlagState(ctx, flagsMap)
	if err != nil {
		zapLogger.Errorf("HandleEventLessonUpdate lessonUpdateChangeLessonStartTime UpdateLessonAutoCreateFlagState failed, err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
			err, msg.TeacherIdsAfter, msg.LocationIdAfter, msg.StartAtAfter.AsTime())
		return err
	}

	return nil
}

func (s *LessonNatsServiceImpl) mappingStaffAndAutoCreateFlagActivityLog(ctx context.Context, lessonStartDate time.Time, teacherIDs []string) (map[string][]*dto.AutoCreateFlagActivityLog, error) {
	mapAutoCreateFlagLog := make(map[string][]*dto.AutoCreateFlagActivityLog)

	autoCreateFlagActivityLog, err := s.AutoCreateFlagActivityLogService.GetAutoCreateFlagLogByTeacherIDs(ctx, lessonStartDate, teacherIDs)
	if err != nil {
		return nil, err
	}

	for _, autoActivityLog := range autoCreateFlagActivityLog {
		mapAutoCreateFlagLog[autoActivityLog.StaffID] = append(mapAutoCreateFlagLog[autoActivityLog.StaffID], autoActivityLog)
	}

	return mapAutoCreateFlagLog, nil
}

func (s *LessonNatsServiceImpl) createAndRemoveTimesheetWhenLessonChanged(
	ctx context.Context,
	lessonID string,
	teacherBef, teacherAft []string,
	locationBef, locationAft string,
	dateBef, dateAft time.Time,
) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	var (
		newTimesheetDatas             []*dto.Timesheet
		removeTimesheetDatas          []*dto.Timesheet
		actionLogTimesheetRecipients  []*dto.Timesheet
		partnerAutoCreateDefaultValue = constant.KPartnerAutoCreateDefaultValue
	)

	isAllowToModifyBeforeUpdate, err := s.ConfirmationWindowService.CheckModifyConditionByTimesheetDateAndLocation(ctx, &timestamppb.Timestamp{Seconds: dateBef.Unix()}, locationBef)
	if err != nil {
		zapLogger.Errorf("HandleEventLessonUpdate CheckModifyConditionByTimesheetDateAndLocation Before Update failed, err: %v, locationID: %v, timesheetDate: %v",
			err, locationBef, dateBef)
		return err
	}
	isAllowToModifyAfterUpdate, err := s.ConfirmationWindowService.CheckModifyConditionByTimesheetDateAndLocation(ctx, &timestamppb.Timestamp{Seconds: dateAft.Unix()}, locationAft)
	if err != nil {
		zapLogger.Errorf("HandleEventLessonUpdate CheckModifyConditionByTimesheetDateAndLocation After Update failed, err: %v, locationID: %v, timesheetDate: %v",
			err, locationAft, dateAft)
		return err
	}

	// build remove before
	if len(teacherBef) > 0 && isAllowToModifyBeforeUpdate {
		needRemoveTimesheets, err := s.GetTimesheetService.GetTimesheet(ctx,
			&dto.TimesheetQueryArgs{
				StaffIDs:      teacherBef,
				LocationID:    locationBef,
				TimesheetDate: dateBef,
			},
			&dto.TimesheetGetOptions{
				IsGetListOtherWorkingHours:     true,
				IsGetListTimesheetLessonHours:  true,
				IsGetListTransportationExpense: true,
			},
		)

		if err != nil {
			zapLogger.Errorf("createAndRemoveTimesheetWhenLessonChanged GetTimesheet in build location before err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
				err, teacherBef, locationBef, dateBef)
			return err
		}
		removeTimesheetDatas = removeTimesheetLessonHours(ctx, lessonID, needRemoveTimesheets)
		// add removeTimesheetDatas to actionLogTimesheetRecipients, timesheets that are to be deleted are ignored when publishing action logs
		actionLogTimesheetRecipients = append(actionLogTimesheetRecipients, removeTimesheetDatas...)
	}

	// build add after
	if len(teacherAft) > 0 && isAllowToModifyAfterUpdate {
		existedTimesheets, err := s.GetTimesheetService.GetTimesheet(ctx,
			&dto.TimesheetQueryArgs{
				StaffIDs:      teacherAft,
				LocationID:    locationAft,
				TimesheetDate: dateAft,
			},
			&dto.TimesheetGetOptions{
				IsGetListTimesheetLessonHours: true,
			},
		)

		if err != nil {
			zapLogger.Errorf("createAndRemoveTimesheetWhenLessonChanged GetTimesheet in build after err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
				err, teacherAft, locationAft, dateAft)
			return err
		}

		// TODO: Move this block of code to common function to reuse for other places
		// Create a map to combine teacherID and AutoCreateFlagActivityLog
		mapAutoCreateFlagLog := make(map[string][]*dto.AutoCreateFlagActivityLog)
		autoCreateFlagActivityLog, err := s.AutoCreateFlagActivityLogService.GetAutoCreateFlagLogByTeacherIDs(ctx, dateAft, teacherAft)
		if err != nil {
			zapLogger.Errorf("createAndRemoveTimesheetWhenLessonChanged GetAutoCreateFlagLogByTeacherIDs after err: %v, staffIDs: %v, locationID: %v, timesheetDate: %v",
				err, teacherAft, locationAft, dateAft)
			return err
		}

		for _, autoActivityLog := range autoCreateFlagActivityLog {
			mapAutoCreateFlagLog[autoActivityLog.StaffID] = append(mapAutoCreateFlagLog[autoActivityLog.StaffID], autoActivityLog)
		}

		partnerAutoCreateDefaultE, err := s.PartnerAutoCreateTimesheetRepo.GetPartnerAutoCreateDefaultValue(ctx, s.DB)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			zapLogger.Errorf("HandleEventLessonCreate GetPartnerAutoCreateDefaultValue failed, err: %v", err)
			return err
		}

		if partnerAutoCreateDefaultE != nil {
			partnerAutoCreateDefaultValue = partnerAutoCreateDefaultE.FlagOn.Bool
		}

		newTimesheetDatas = buildTimesheetData(
			lessonID,
			locationAft,
			teacherAft,
			dateAft,
			existedTimesheets,
			mapAutoCreateFlagLog,
			partnerAutoCreateDefaultValue,
		)
		// add newTimesheetDatas to actionLogTimesheetRecipients, timesheets that are to be created are ignored when publishing action logs
		actionLogTimesheetRecipients = append(actionLogTimesheetRecipients, newTimesheetDatas...)
	}

	// create and remove timesheet
	err = s.AutoCreateTimesheetService.CreateAndRemoveTimesheetMultiple(ctx, newTimesheetDatas, removeTimesheetDatas)
	if err != nil {
		zapLogger.Errorf("createAndRemoveTimesheetWhenLessonChanged CreateAndRemoveTimesheetMultiple failed, err: %v, lessonID: %v", err, lessonID)
		return err
	}

	if err = s.publishSystemUpdateLessonActionLogEvent(ctx, actionLogTimesheetRecipients); err != nil {
		zapLogger.Errorf("createAndRemoveTimesheetWhenLessonChanged publishSystemUpdateLessonActionLogEvent failed, err: %v, lessonID: %v, params: %v", err, lessonID, actionLogTimesheetRecipients)
		return err
	}

	return nil
}

func (s *LessonNatsServiceImpl) publishSystemUpdateLessonActionLogEvent(ctx context.Context, timesheets []*dto.Timesheet) error {
	timeExecuted := time.Now()
	for _, timesheet := range timesheets {
		// if timesheet isn't created yet or is deleted, don't send action log event
		if !timesheet.IsCreated || timesheet.IsDeleted {
			continue
		}

		// send action log event
		msg := &tpb.TimesheetActionLogRequest{
			Action:      tpb.TimesheetAction_UPDATED_LESSON,
			ExecutedBy:  interceptors.UserIDFromContext(ctx),
			TimesheetId: timesheet.ID,
			IsSystem:    true,
			ExecutedAt:  timestamppb.New(timeExecuted),
		}
		err := PublishActionLogTimesheetEvent(ctx, msg, s.JSM)
		if err != nil {
			return err
		}
	}

	return nil
}
