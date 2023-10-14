package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	v11 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"google.golang.org/protobuf/proto"
)

func checkLessonUpdateType(msg *bpb.EvtLesson_UpdateLesson) []constant.LessonUpdateType {
	var updateType []constant.LessonUpdateType

	if isScheduleStatusFromPublishedToDraft(msg) {
		updateType = append(updateType, constant.LessonUpdateTypePublishedToDraft)
	}

	if isScheduleStatusFromDraftToPublished(msg) {
		updateType = append(updateType, constant.LessonUpdateTypeDraftToPublished)
	}

	if isChangedLessonHoursTeacher(msg) {
		updateType = append(updateType, constant.LessonUpdateTypeChangeTeacherID)
	}

	if isLessonHoursDateChanged(msg) {
		updateType = append(updateType, constant.LessonUpdateTypeChangeLessonDate)
	} else if isLessonHoursStartTimeChanged(msg) {
		updateType = append(updateType, constant.LessonUpdateTypeChangeLessonStartTime)
	}

	if isLessonHoursLocationChanged(msg) {
		updateType = append(updateType, constant.LessonUpdateTypeChangeLocationID)
	}

	return updateType
}

func isScheduleStatusFromDraftToPublished(msg *bpb.EvtLesson_UpdateLesson) bool {
	if msg.SchedulingStatusBefore == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT &&
		msg.SchedulingStatusAfter == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED {
		return true
	}
	return false
}

func isScheduleStatusFromPublishedToDraft(msg *bpb.EvtLesson_UpdateLesson) bool {
	if msg.SchedulingStatusBefore == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED &&
		msg.SchedulingStatusAfter == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
		return true
	}
	return false
}

func isChangedLessonHoursTeacher(msg *bpb.EvtLesson_UpdateLesson) bool {
	if msg.SchedulingStatusAfter == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
		return false
	}
	return !stringutil.SliceEqual(msg.TeacherIdsBefore, msg.TeacherIdsAfter)
}

func isLessonHoursDateChanged(msg *bpb.EvtLesson_UpdateLesson) bool {
	if msg.SchedulingStatusAfter == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
		return false
	}
	return !timeutil.EqualDate(msg.GetStartAtBefore().AsTime().In(timeutil.Timezone(pbc.COUNTRY_JP)), msg.GetStartAtAfter().AsTime().In(timeutil.Timezone(pbc.COUNTRY_JP)))
}

func isLessonHoursStartTimeChanged(msg *bpb.EvtLesson_UpdateLesson) bool {
	if msg.SchedulingStatusAfter == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
		return false
	}
	return !msg.GetStartAtBefore().AsTime().Equal(msg.GetStartAtAfter().AsTime())
}

func isLessonHoursLocationChanged(msg *bpb.EvtLesson_UpdateLesson) bool {
	if msg.SchedulingStatusAfter == v11.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
		return false
	}
	return msg.GetLocationIdBefore() != msg.GetLocationIdAfter()
}

/*
if timesheet does not exist create a new timesheet with lesson hours
else add new lesson hours into the timesheet existing
return list timesheet with up to date timesheet data and timesheet lesson hours data
*/
func buildTimesheetData(lessonID, locationID string, teacherIDs []string, lessonStartDate time.Time, timesheets []*dto.Timesheet, mapStaffIDAutoCreateFlagLog map[string][]*dto.AutoCreateFlagActivityLog, partnerAutoCreateDefaultValue bool) []*dto.Timesheet {
	mapStaffIDTimesheetID := getMapStaffIDTimesheet(teacherIDs, timesheets)

	for _, teacherID := range teacherIDs {
		timesheet := mapStaffIDTimesheetID[teacherID]
		if timesheet == nil {
			timesheet = buildTimesheetFromEvtLessonUpdateLesson(lessonID, locationID, teacherID, lessonStartDate, mapStaffIDAutoCreateFlagLog[teacherID], partnerAutoCreateDefaultValue)
			mapStaffIDTimesheetID[teacherID] = timesheet
		} else {
			timesheet.ListTimesheetLessonHours = append(
				timesheet.ListTimesheetLessonHours,
				buildTimesheetLessonHour(lessonID, timesheet.ID, mapStaffIDAutoCreateFlagLog[teacherID], partnerAutoCreateDefaultValue))
		}
	}

	rsTimesheets := make([]*dto.Timesheet, 0, len(mapStaffIDTimesheetID))
	for _, value := range mapStaffIDTimesheetID {
		if isValidateTimesheetStatus(value) && value.TimesheetDate.After(constant.KTimesheetMinDate) {
			rsTimesheets = append(rsTimesheets, value)
		}
	}
	return rsTimesheets
}

// return list timesheet base on teacherIDs
func buildTimesheetDataFromLesson(msg *bpb.EvtLesson_Lesson, timesheets []*dto.Timesheet, mapStaffIDAutoCreateFlagLog map[string][]*dto.AutoCreateFlagActivityLog, partnerAutoCreateDefault bool) []*dto.Timesheet {
	mapStaffIDTimesheetID := getMapStaffIDTimesheet(msg.TeacherIds, timesheets)

	for _, teacherID := range msg.TeacherIds {
		timesheet := mapStaffIDTimesheetID[teacherID]
		if timesheet == nil {
			timesheet = buildTimesheetFromEvtLessonLessonCreate(msg, teacherID, mapStaffIDAutoCreateFlagLog[teacherID], partnerAutoCreateDefault)
			mapStaffIDTimesheetID[teacherID] = timesheet
		} else {
			timesheet.ListTimesheetLessonHours = append(
				timesheet.ListTimesheetLessonHours,
				buildTimesheetLessonHour(msg.LessonId, timesheet.ID, mapStaffIDAutoCreateFlagLog[teacherID], partnerAutoCreateDefault))
		}
	}

	rsTimesheets := make([]*dto.Timesheet, 0, len(mapStaffIDTimesheetID))
	for _, value := range mapStaffIDTimesheetID {
		if isValidateTimesheetStatus(value) && value.TimesheetDate.After(constant.KTimesheetMinDate) {
			rsTimesheets = append(rsTimesheets, value)
		}
	}
	return rsTimesheets
}

func buildTimesheetLessonHour(lessonID string, timesheetID string, listOfAutoCreateFlagLog []*dto.AutoCreateFlagActivityLog, partnerAutoCreateDefault bool) *dto.TimesheetLessonHours {
	var latestAutoCreateFlagLog bool
	if len(listOfAutoCreateFlagLog) > 0 {
		latestAutoCreateFlagLog = listOfAutoCreateFlagLog[0].FlagOn
	} else {
		latestAutoCreateFlagLog = partnerAutoCreateDefault
	}

	return &dto.TimesheetLessonHours{
		TimesheetID: timesheetID,
		LessonID:    lessonID,
		FlagOn:      latestAutoCreateFlagLog,
	}
}

func buildTimesheetFromEvtLessonUpdateLesson(lessonID, locationID, teacherID string, lessonStartDate time.Time, listOfAutoCreateFlagLog []*dto.AutoCreateFlagActivityLog, partnerAutoCreateDefaultValue bool) *dto.Timesheet {
	var latestAutoCreateFlagLog bool
	if len(listOfAutoCreateFlagLog) > 0 {
		latestAutoCreateFlagLog = listOfAutoCreateFlagLog[0].FlagOn
	} else {
		latestAutoCreateFlagLog = partnerAutoCreateDefaultValue
	}

	listTimesheetLessonHours := make(dto.ListTimesheetLessonHours, 0)
	listTimesheetLessonHours = append(
		listTimesheetLessonHours,
		&dto.TimesheetLessonHours{LessonID: lessonID, FlagOn: latestAutoCreateFlagLog},
	)

	return &dto.Timesheet{
		StaffID:                  teacherID,
		LocationID:               locationID,
		TimesheetStatus:          pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:            lessonStartDate,
		ListTimesheetLessonHours: listTimesheetLessonHours,
		IsCreated:                false,
	}
}

func buildTimesheetLessonAutoCreateFlagMap(timesheets []*dto.Timesheet, mapStaffIDAutoCreateFlagLog map[string][]*dto.AutoCreateFlagActivityLog, partnerAutoCreateDefaultValue bool) map[bool][]string {
	mapFlags := map[bool][]string{}
	for _, ts := range timesheets {
		flagOn := false
		val, ok := mapStaffIDAutoCreateFlagLog[ts.StaffID]
		if ok && len(val) > 0 {
			flagOn = val[0].FlagOn
		} else {
			flagOn = partnerAutoCreateDefaultValue
		}

		mapFlags[flagOn] = append(mapFlags[flagOn], ts.ID)
	}

	return mapFlags
}

func getMapStaffIDTimesheet(teacherIDs []string, timesheets []*dto.Timesheet) map[string]*dto.Timesheet {
	mapStaffIDTimesheet := make(map[string]*dto.Timesheet, len(teacherIDs))
	mapStaffIDTimesheetExists := make(map[string]*dto.Timesheet, len(timesheets))

	for _, timesheet := range timesheets {
		mapStaffIDTimesheetExists[timesheet.StaffID] = timesheet
	}

	for _, teacherID := range teacherIDs {
		mapStaffIDTimesheet[teacherID] = mapStaffIDTimesheetExists[teacherID]
	}

	return mapStaffIDTimesheet
}

func buildTimesheetFromEvtLessonLessonCreate(msg *bpb.EvtLesson_Lesson, teacherID string, listOfAutoCreateFlagLog []*dto.AutoCreateFlagActivityLog, partnerAutoCreateDefault bool) *dto.Timesheet {
	var latestAutoCreateFlagLog bool
	if len(listOfAutoCreateFlagLog) > 0 {
		latestAutoCreateFlagLog = listOfAutoCreateFlagLog[0].FlagOn
	} else {
		latestAutoCreateFlagLog = partnerAutoCreateDefault
	}

	listTimesheetLessonHours := make(dto.ListTimesheetLessonHours, 0)
	listTimesheetLessonHours = append(listTimesheetLessonHours,
		&dto.TimesheetLessonHours{
			LessonID: msg.LessonId,
			FlagOn:   latestAutoCreateFlagLog,
		})

	return &dto.Timesheet{
		StaffID:                  teacherID,
		LocationID:               msg.LocationId,
		TimesheetStatus:          pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		TimesheetDate:            msg.StartAt.AsTime(),
		ListTimesheetLessonHours: listTimesheetLessonHours,
	}
}

func removeTimesheetLessonHours(ctx context.Context, lessonID string, timesheets []*dto.Timesheet) []*dto.Timesheet {
	var (
		zapLogger = ctxzap.Extract(ctx).Sugar()
	)
	removedTimesheets := make([]*dto.Timesheet, 0, len(timesheets))
	for _, timesheet := range timesheets {
		if !isValidateTimesheetStatus(timesheet) {
			continue
		}

		isTimesheetLessonHoursDeleted := timesheet.DeleteTimesheetLessonHours(lessonID)

		if isTimesheetLessonHoursDeleted {
			if timesheet.IsTimesheetEmpty() {
				timesheet.IsDeleted = true
			}
		} else {
			zapLogger.Warnf("Lesson with ID=%v not belong to timesheet=%v", lessonID, timesheet.ID)
		}
		removedTimesheets = append(removedTimesheets, timesheet)
	}
	return removedTimesheets
}

func isValidateTimesheetStatus(timesheet *dto.Timesheet) bool {
	if timesheet.TimesheetStatus != pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String() &&
		timesheet.TimesheetStatus != pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String() {
		return false
	}
	return true
}

func calculateLessonTeachersChanged(before, after []string) (new, removed []string) {
	removed = stringutil.SliceElementsDiff(before, after)
	new = stringutil.SliceElementsDiff(after, before)
	return new, removed
}

func PublishActionLogTimesheetEvent(ctx context.Context, msg *pb.TimesheetActionLogRequest, jsm nats.JetStreamManagement) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgID, err := jsm.PublishAsyncContext(ctx, constants.SubjectTimesheetActionLog, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishActionLogTimesheetEvent JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}

func PublishUpdateAutoCreateFlagEvent(ctx context.Context, msg *pb.NatsUpdateAutoCreateTimesheetFlagRequest, jsm nats.JetStreamManagement) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgID, err := jsm.PublishAsyncContext(ctx, constants.SubjectTimesheetAutoCreateFlag, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishUpdateAutoCreateFlagEvent JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}
