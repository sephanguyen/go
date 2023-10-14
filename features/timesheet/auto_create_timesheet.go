package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	"github.com/manabie-com/backend/internal/timesheet/infrastructure/repository"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) checkTimesheetIsCreated(ctx context.Context, total int, isCreated string) (context.Context, error) {
	time.Sleep(3 * time.Second) // wait service process event from nats
	stepState := StepStateFromContext(ctx)
	var (
		locationID    string
		timesheetDate time.Time
	)

	if stepState.Lesson != nil {
		locationID = stepState.Lesson.LocationID
		timesheetDate = stepState.Lesson.StartTime
	} else if stepState.Request != nil {
		switch v := stepState.Request.(type) {
		case *bpb.CreateLessonRequest:
			locationID = v.CenterId
			timesheetDate = v.StartTime.AsTime()
		case *bpb.UpdateLessonRequest:
			locationID = v.CenterId
			timesheetDate = v.StartTime.AsTime()
		case *lpb.CreateLessonRequest:
			locationID = v.LocationId
			timesheetDate = v.StartTime.AsTime()
		case *lpb.UpdateLessonRequest:
			locationID = v.LocationId
			timesheetDate = v.StartTime.AsTime()
		}
	} else {
		return ctx, fmt.Errorf("invalid stepState, not found locationID, timesheetDate")
	}

	timesheetArgs := &dto.TimesheetQueryArgs{
		StaffIDs:      stepState.TeacherIDs,
		LocationID:    locationID,
		TimesheetDate: timesheetDate,
	}
	timesheetArgs.Normalize()
	timesheetRepo := &repository.TimesheetRepoImpl{}
	timesheets, err := timesheetRepo.FindTimesheetByTimesheetArgs(ctx, s.TimesheetDB, timesheetArgs)
	if err != nil {
		return ctx, err
	}

	if isCreated == "created" && len(timesheets) == total {
		for _, timesheet := range timesheets {
			stepState.CurrentTimesheetIDs = append(stepState.CurrentTimesheetIDs, timesheet.TimesheetID.String)
			stepState.Timesheets = append(stepState.Timesheets, timesheet)
		}
		return StepStateToContext(ctx, stepState), nil
	}

	if isCreated == "not created" && len(timesheets) == total {
		return StepStateToContext(ctx, stepState), nil
	}

	if isCreated == "exists" && len(timesheets) == total {
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), fmt.Errorf("timesheet create not as expected, isCreated: %v,total: %d", isCreated, len(timesheets))
}

func (s *Suite) checkTimesheetsChangedWhenLessonDateChanged(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second) // wait service process event from nats
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*lpb.UpdateLessonRequest)

	// Get timesheets with old lesson date
	timesheetArgs := &dto.TimesheetQueryArgs{
		StaffIDs:      req.TeacherIds,
		LocationID:    req.LocationId,
		TimesheetDate: stepState.TimesheetDateBeforeChange,
	}
	timesheetArgs.Normalize()
	timesheetRepo := &repository.TimesheetRepoImpl{}
	oldTimesheets, err := timesheetRepo.FindTimesheetByTimesheetArgs(ctx, s.TimesheetDB, timesheetArgs)
	if err != nil {
		return ctx, err
	}

	if len(oldTimesheets) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("timesheet with old lesson date was not removed as expected")
	}

	// Get timesheet with new lesson date
	timesheetArgs2 := &dto.TimesheetQueryArgs{
		StaffIDs:      req.TeacherIds,
		LocationID:    req.LocationId,
		TimesheetDate: req.StartTime.AsTime(),
	}
	timesheetArgs2.Normalize()
	currentTimesheets, err := timesheetRepo.FindTimesheetByTimesheetArgs(ctx, s.TimesheetDB, timesheetArgs2)
	if err != nil {
		return ctx, err
	}
	if len(currentTimesheets) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("timesheet with new lesson date was not added, expected teacherIDs:%v,locationID:%v, startTime:%v", req.TeacherIds, req.LocationId, req.StartTime.AsTime())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkTimesheetsChangedWhenLessonLocationChanged(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second) // wait service process event from nats
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*lpb.UpdateLessonRequest)

	// Get timesheets with old lesson location
	timesheetArgs := &dto.TimesheetQueryArgs{
		StaffIDs:      req.TeacherIds,
		LocationID:    stepState.TimesheetLocationBeforeChange,
		TimesheetDate: req.StartTime.AsTime(),
	}
	timesheetArgs.Normalize()
	timesheetRepo := &repository.TimesheetRepoImpl{}
	oldTimesheets, err := timesheetRepo.FindTimesheetByTimesheetArgs(ctx, s.TimesheetDB, timesheetArgs)
	if err != nil {
		return ctx, err
	}

	if len(oldTimesheets) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("timesheet with old lesson location was not removed as expected")
	}

	// Get timesheet with new lesson location
	timesheetArgs2 := &dto.TimesheetQueryArgs{
		StaffIDs:      req.TeacherIds,
		LocationID:    req.LocationId,
		TimesheetDate: req.StartTime.AsTime(),
	}
	timesheetArgs2.Normalize()
	currentTimesheets, err := timesheetRepo.FindTimesheetByTimesheetArgs(ctx, s.TimesheetDB, timesheetArgs2)
	if err != nil {
		return ctx, err
	}
	if len(currentTimesheets) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("timesheet with new lesson location was not added, expected teacherIDs:%v,locationID:%v, startTime:%v", req.TeacherIds, req.LocationId, req.StartTime.AsTime())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkTimesheetLessonHoursIsCreated(ctx context.Context, total int, isCreated string) (context.Context, error) {
	time.Sleep(10 * time.Second) // wait service process event from nats
	stepState := StepStateFromContext(ctx)

	timesheetLessonHoursRepo := &repository.TimesheetLessonHoursRepoImpl{}
	timesheetLessonHours, err := timesheetLessonHoursRepo.FindByTimesheetIDs(ctx, s.TimesheetDB, stepState.CurrentTimesheetIDs)
	if err != nil {
		return ctx, err
	}
	if isCreated == "created" && len(timesheetLessonHours) == total {
		for _, e := range timesheetLessonHours {
			stepState.CurrentListTimesheetLessonHours = append(stepState.CurrentListTimesheetLessonHours, dto.NewTimesheetLessonHoursFromEntity(e))
		}
		return StepStateToContext(ctx, stepState), nil
	}

	if isCreated == "not created" && len(timesheetLessonHours) == total {
		return StepStateToContext(ctx, stepState), nil
	}
	return StepStateToContext(ctx, stepState), fmt.Errorf("timesheetLessonHours create not as expected, isCreated: %v, total: %d", isCreated, len(timesheetLessonHours))
}
func (s *Suite) checkTimesheetLessonHoursIsValid(ctx context.Context, valid string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var flagOn bool
	if valid == "on" {
		flagOn = true
	}
	for _, timesheetLessonHours := range stepState.CurrentListTimesheetLessonHours {
		if timesheetLessonHours.FlagOn != flagOn {
			return StepStateToContext(ctx, stepState), fmt.Errorf("timesheetLessonHours have flagOn not as expected, timesheet id: %v, lesson id: %v, flagOn: %v, expect flagOn: %v", timesheetLessonHours.TimesheetID, timesheetLessonHours.LessonID, timesheetLessonHours.FlagOn, valid)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *Suite) UserCreateALessonWithAllRequiredFieldsWithSub(ctx context.Context) (context.Context, error) {
	return s.UserCreateALessonWithAllRequiredFields(ctx)
}

func (s *Suite) UserCreateALessonWithAllRequiredFields(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"
	ctx, _ = s.CommonSuite.UserCreateALessonWithMissingFields(ctx)
	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *Suite) UserCreateALessonWithAllRequiredFieldsInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"
	ctx, _ = s.CommonSuite.UserCreateALessonWithMissingFieldsInLessonmgmt(ctx)
	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}

func (s *Suite) userUpdatesStatusInTheLessonIsValue(ctx context.Context, value string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var lessonStatus lpb.LessonStatus
	switch value {
	case cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT.String():
		lessonStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
	case cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String():
		lessonStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("LessonSchedulingStatus is not available")
	}

	var updateLessonReq *lpb.UpdateLessonRequest
	switch req := stepState.Request.(type) {
	case *lpb.CreateLessonRequest:
		studentInfos := []*lpb.UpdateLessonRequest_StudentInfo{}
		for _, studentInf := range req.StudentInfoList {
			studentInfos = append(studentInfos, &lpb.UpdateLessonRequest_StudentInfo{
				StudentId:        studentInf.StudentId,
				CourseId:         studentInf.CourseId,
				AttendanceStatus: studentInf.AttendanceStatus,
			})
		}

		updateLessonReq = &lpb.UpdateLessonRequest{
			LessonId:         stepState.CurrentLessonID,
			StartTime:        req.StartTime,
			EndTime:          req.EndTime,
			TeachingMedium:   req.TeachingMedium,
			TeachingMethod:   req.TeachingMethod,
			LocationId:       req.LocationId,
			SchedulingStatus: lessonStatus,
			TeacherIds:       req.TeacherIds,
			StudentInfoList:  studentInfos,
			SavingOption: &lpb.UpdateLessonRequest_SavingOption{
				Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
			},
		}

	case *lpb.UpdateLessonRequest:
		updateLessonReq = req
		updateLessonReq.SchedulingStatus = lessonStatus
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected type, got %T", stepState.Request)
	}

	stepState.CurrentLessonID = updateLessonReq.LessonId

	ctx, _ = s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateLessonReq)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("LessonSchedulingStatus update lesson got error:%v", stepState.ResponseErr)
	}
	return ctx, nil
}

func (s *Suite) userUpdateTeacherInTheLesson(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	updateRequest := updateRequestFromCreateRequestInLessonmgmt(stepState.CurrentLessonID, createdRequest)

	// have 2 teacher updateRequest.TeacherIds[0] , updateRequest.TeacherIds[1]
	switch action {
	case "remove": // remove updateRequest.TeacherIds[1]
		updateRequest.TeacherIds = []string{updateRequest.TeacherIds[0]}
	case "add": // add stepState.TeacherIDsUpdateLesson[0]
		updateRequest.TeacherIds = append(updateRequest.TeacherIds, stepState.TeacherIDsUpdateLesson[0])
	case "2 remove and 1 add": // remove updateRequest.TeacherIds[0], updateRequest.TeacherIds[1], add stepState.TeacherIDsUpdateLesson[1]
		updateRequest.TeacherIds = []string{stepState.TeacherIDsUpdateLesson[1]}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong update lesson teacher action, got %s", action)
	}
	stepState.TeacherIDs = updateRequest.TeacherIds
	ctx, _ = s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
	stepState = StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("UserUpdateALessonWithRequestInLessonmgmt, err %s", stepState.ResponseErr.Error())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateNewTeacherInTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	updateRequest := updateRequestFromCreateRequestInLessonmgmt(stepState.CurrentLessonID, createdRequest)

	updateRequest.TeacherIds = []string{stepState.TeacherIDsUpdateLesson[0]}
	ctx, _ = s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
	stepState = StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("UserUpdateALessonWithRequestInLessonmgmt, err %s", stepState.ResponseErr.Error())
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, updateRequest.TeacherIds...)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateLessonDate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}

	updateRequest := updateRequestFromCreateRequestInLessonmgmt(stepState.CurrentLessonID, createdRequest)

	stepState.TimesheetDateBeforeChange = updateRequest.StartTime.AsTime()

	updateRequest.StartTime = timestamppb.New(updateRequest.StartTime.AsTime().AddDate(0, 0, 5))
	updateRequest.EndTime = timestamppb.New(updateRequest.EndTime.AsTime().AddDate(0, 0, 5))

	ctx, _ = s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
	stepState = StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("UserUpdateALessonWithRequestInLessonmgmt, err %s", stepState.ResponseErr.Error())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateLessonLocation(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}

	updateRequest := updateRequestFromCreateRequestInLessonmgmt(stepState.CurrentLessonID, createdRequest)

	stepState.TimesheetLocationBeforeChange = updateRequest.LocationId

	updateRequest.LocationId = stepState.CenterIDs[len(stepState.CenterIDs)-2]

	ctx, _ = s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
	stepState = StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("UserUpdateALessonWithRequestInLessonmgmt, err %s", stepState.ResponseErr.Error())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheLessonSchedulingStatusWasUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	updatedRequest, ok := stepState.Request.(*lpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.UpdateLessonRequest, got %T", updatedRequest)
	}

	if lesson.SchedulingStatus != domain.LessonSchedulingStatus(updatedRequest.SchedulingStatus.String()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected SchedulingStatus %s but got %s", domain.LessonSchedulingStatus(updatedRequest.SchedulingStatus.String()), lesson.SchedulingStatus)
	}
	stepState.Lesson = lesson
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheLessonTeacherWasUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	updatedRequest, ok := stepState.Request.(*lpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLessonRequest, got %T", updatedRequest)
	}

	if !stringutil.SliceElementsMatch(lesson.GetTeacherIDs(), updatedRequest.TeacherIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected Teachers %v but got %v", updatedRequest.TeacherIds, lesson.GetTeacherIDs())
	}
	stepState.Lesson = lesson
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheLessonFieldWasUpdated(ctx context.Context, field string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	updatedRequest, ok := stepState.Request.(*lpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLessonRequest, got %T", updatedRequest)
	}
	switch field {
	case "date":
		if !lesson.StartTime.Equal(updatedRequest.StartTime.AsTime()) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson date %v but got %v", updatedRequest.StartTime.AsTime(), lesson.StartTime)
		}
	case "TeacherIDs":
		if !stringutil.SliceEqual(lesson.GetTeacherIDs(), updatedRequest.TeacherIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected Teachers %v but got %v", updatedRequest.TeacherIds, lesson.GetTeacherIDs())
		}
	case "location":
		if lesson.LocationID != updatedRequest.LocationId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected locationID %v but got %v", updatedRequest.LocationId, lesson.LocationID)
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected lesson field, got %T", field)
	}

	stepState.Lesson = lesson
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkTimesheetStatus(ctx context.Context, timesheetStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.Timesheets) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("timesheets is empty")
	}
	var validStatus string
	switch timesheetStatus {
	case "Draft":
		validStatus = pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()
	case "Summited":
		validStatus = pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()
	case "Approved":
		validStatus = pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()
	case "Rejected":
		validStatus = pb.TimesheetStatus_TIMESHEET_STATUS_REJECTED.String()
	case "Confirmed":
		validStatus = pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("invalid timesheet status: %v", timesheetStatus)
	}
	for _, e := range stepState.Timesheets {
		if e.TimesheetStatus.String != validStatus {
			return StepStateToContext(ctx, stepState), fmt.Errorf("check timesheet status failed, required: %v", timesheetStatus)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func updateRequestFromCreateRequestInLessonmgmt(lessonID string, createdRequest *lpb.CreateLessonRequest) *lpb.UpdateLessonRequest {
	updateRequest := &lpb.UpdateLessonRequest{
		LessonId:        lessonID,
		StartTime:       createdRequest.StartTime,
		EndTime:         createdRequest.EndTime,
		TeachingMedium:  createdRequest.TeachingMedium,
		TeachingMethod:  createdRequest.TeachingMethod,
		TeacherIds:      createdRequest.TeacherIds,
		LocationId:      createdRequest.LocationId,
		StudentInfoList: make([]*lpb.UpdateLessonRequest_StudentInfo, 0, len(createdRequest.StudentInfoList)),
		Materials:       createdRequest.Materials,
		SavingOption: &lpb.UpdateLessonRequest_SavingOption{
			Method: createdRequest.SavingOption.Method,
		},
	}
	for _, studentInf := range createdRequest.StudentInfoList {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &lpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentInf.StudentId,
			CourseId:         studentInf.CourseId,
			AttendanceStatus: studentInf.AttendanceStatus,
		})
	}

	return updateRequest
}

func (s *Suite) checkTimesheetLessonHoursFlagON(ctx context.Context, flag string) (context.Context, error) {
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

	for _, tlh := range timesheetLessonHours {
		if tlh.FlagOn.Bool != flagOn {
			return StepStateToContext(ctx, stepState), fmt.Errorf(
				"timesheetLessonHours update not as expected, timesheetID:%s, flagOn:%t",
				tlh.TimesheetID.String, tlh.FlagOn.Bool)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
