package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

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
		LessonCapacity: 45,
	}
	for _, studentInf := range createdRequest.StudentInfoList {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &lpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentInf.StudentId,
			CourseId:         studentInf.CourseId,
			AttendanceStatus: studentInf.AttendanceStatus,
			AttendanceNote:   studentInf.AttendanceNote,
			AttendanceNotice: studentInf.AttendanceNotice,
			AttendanceReason: studentInf.AttendanceReason,
			LocationId:       createdRequest.LocationId,
		})
	}

	return updateRequest
}

//nolint:gocyclo
func (s *Suite) userUpdatesFieldInTheLessonLessonmgmt(ctx context.Context, fieldName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randID := idutil.ULIDNow()
	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	updateRequest := updateRequestFromCreateRequestInLessonmgmt(stepState.CurrentLessonID, createdRequest)

	// Change the field to something else for the update request
	var err error
	updateAllFields := fieldName == "all fields"
	if updateAllFields || fieldName == "start time" || fieldName == "starttime" || fieldName == "start_time" {
		updateRequest.StartTime = timestamppb.New(updateRequest.StartTime.AsTime().Add(time.Hour))
	}
	if updateAllFields || fieldName == "end time" || fieldName == "endtime" || fieldName == "end_time" {
		updateRequest.EndTime = timestamppb.New(updateRequest.EndTime.AsTime().Add(time.Hour))
	}
	if updateAllFields || fieldName == "center id" || fieldName == "centerid" || fieldName == "center_id" {
		updateRequest.LocationId = constants.ManabieOrgLocation
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for i := range updateRequest.StudentInfoList {
			updateRequest.StudentInfoList[i].LocationId = constants.ManabieOrgLocation
		}
	}
	if updateAllFields || fieldName == "teacher ids" || fieldName == "teacherids" || fieldName == "teacher_ids" {
		updateRequest.TeacherIds = updateRequest.TeacherIds[:len(updateRequest.TeacherIds)-1]
		stepState.TeacherIDs, err = s.updateLessonTeacherIDs(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		updateRequest.TeacherIds = append(updateRequest.TeacherIds, stepState.TeacherIDs...)
	}
	if updateAllFields || fieldName == "student info list" || fieldName == "studentinfolist" || fieldName == "student_info_list" {
		updateRequest.StudentInfoList = updateRequest.StudentInfoList[:len(updateRequest.StudentInfoList)-1]
		studentInfoList, err := s.updateStudentInfoLessonmgmt(ctx, updateRequest.LocationId)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, studentInfoList...)
	}
	if updateAllFields || fieldName == "teaching medium" || fieldName == "teachingmedium" || fieldName == "teaching_medium" {
		if updateRequest.TeachingMedium == cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE {
			updateRequest.TeachingMedium = cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ONLINE
		} else {
			updateRequest.TeachingMedium = cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE
		}
	}
	if stepState.CurrentTeachingMethod == "group" {
		updateRequest.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP
		updateRequest.CourseId = "bdd_test_update_lesson_course_id_" + randID
		updateRequest.ClassId = "bdd_test_update_lesson_class_id_" + randID
		// update for comparisons
		stepState.CurrentCourseID = updateRequest.CourseId
		stepState.CurrentClassId = updateRequest.ClassId
		// add course & class since lesson domain will check for course & class existed in DB
		if ctx, err := s.createClassAndCourseWithID(ctx, stepState.CurrentClassId, stepState.CurrentCourseID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	if stepState.CurrentTeachingMethod == "individual" {
		updateRequest.TeachingMethod = cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL
	}

	if updateAllFields || fieldName == "material info" || fieldName == "materials" {
		updateRequest.Materials = updateRequest.Materials[:len(updateRequest.Materials)-1]
		materials, err := s.updateMaterialInLessonmgmt(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		updateRequest.Materials = append(updateRequest.Materials, materials...)
	}
	if updateAllFields || fieldName == "lesson capacity" {
		updateRequest.LessonCapacity = 99
	}
	stepState.CurrentLessonID = updateRequest.LessonId
	stepState.Request = updateRequest
	ctx, err = s.CommonSuite.CreateEditLessonSubscriptionLessonmgmt(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditFieldLessonSubscription: %w", err)
	}

	return s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
}

func (s *Suite) userUpdatedLocationAndLessonTime(ctx context.Context, _location, _startDate, _endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, _ := time.Parse(time.RFC3339, _startDate)
	endDate, _ := time.Parse(time.RFC3339, _endDate)

	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	updateLessonReq := updateRequestFromCreateRequestInLessonmgmt(stepState.CurrentLessonID, createdRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("wrong type %T", updateLessonReq)
	}
	locationIndex, err := strconv.Atoi(_location)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("parse location fail %v", err)
	}
	updateLessonReq.LocationId = stepState.CenterIDs[locationIndex]
	updateLessonReq.StartTime = timestamppb.New(startDate)
	updateLessonReq.EndTime = timestamppb.New(endDate)
	stepState.CurrentLessonID = updateLessonReq.LessonId
	stepState.Request = updateLessonReq
	for i := range updateLessonReq.StudentInfoList {
		updateLessonReq.StudentInfoList[i].LocationId = updateLessonReq.LocationId
	}

	ctx, err = s.CommonSuite.CreateEditLessonSubscriptionLessonmgmt(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditFieldLessonSubscription: %w", err)
	}

	return s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateLessonReq)
}

func (s *Suite) updateMaterialInLessonmgmt(ctx context.Context) ([]*lpb.Material, error) {
	stepState := StepStateFromContext(ctx)
	stepState.MediaIDs = []string{}
	_, err := s.CommonSuite.UpsertValidMediaList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create media: %s", err)
	}
	if len(stepState.MediaIDs) < 1 {
		return nil, fmt.Errorf("required at least 1 media, got %d", len(stepState.MediaIDs))
	}

	materials := make([]*lpb.Material, 0, len(stepState.MediaIDs))
	for _, id := range stepState.MediaIDs {
		materials = append(materials,
			&lpb.Material{
				Resource: &lpb.Material_MediaId{
					MediaId: id,
				},
			},
		)
	}
	return materials, nil
}

func (s *Suite) updateStudentInfoLessonmgmt(ctx context.Context, locationID string) ([]*lpb.UpdateLessonRequest_StudentInfo, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentIds = []string{}
	ctx, err := s.CreateStudentAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create students: %s", err)
	}
	if len(stepState.StudentIds) < 1 {
		return nil, fmt.Errorf("required at least 1 student, got %d", len(stepState.StudentIds))
	}
	_, err = s.CommonSuite.SomeStudentSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create student subscription: %s", err)
	}

	studentInfoList := make([]*lpb.UpdateLessonRequest_StudentInfo, 0, len(stepState.StudentIds))
	for _, studentID := range stepState.StudentIds {
		studentInfoList = append(studentInfoList,
			&lpb.UpdateLessonRequest_StudentInfo{
				StudentId:        studentID,
				CourseId:         stepState.CourseIDs[len(stepState.CourseIDs)-1],
				AttendanceStatus: lpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_ABSENT,
				AttendanceNotice: lpb.StudentAttendanceNotice_IN_ADVANCE,
				AttendanceReason: lpb.StudentAttendanceReason_FAMILY_REASON,
				AttendanceNote:   "sample attendance note",
				LocationId:       locationID,
			},
		)
	}

	return studentInfoList, nil
}

func (s *Suite) TheLessonWasUpdatedInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	updatedRequest, ok := stepState.Request.(*lpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.UpdateLiveLessonRequest, got %T", updatedRequest)
	}

	if ctx, err = s.validateLessonForUpdateRequestMGMTInLessonmgmt(ctx, lesson, updatedRequest); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for update lesson: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateALessonWithStartTimeLaterThanEndTimeInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.CreateLessonRequest, got %T", createdRequest)
	}
	updateRequest := &lpb.UpdateLessonRequest{
		LessonId:        stepState.CurrentLessonID,
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
			AttendanceNote:   studentInf.AttendanceNote,
			AttendanceNotice: studentInf.AttendanceNotice,
			AttendanceReason: studentInf.AttendanceReason,
		})
	}

	updateRequest.StartTime = timestamppb.New(updateRequest.StartTime.AsTime().Add(2 * time.Hour))
	updateRequest.EndTime = timestamppb.New(updateRequest.StartTime.AsTime().Add(-time.Hour))
	stepState.Request = updateRequest
	ctx, err := s.CommonSuite.CreateEditLessonSubscriptionLessonmgmt(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditLessonWithStartTimeLaterThanEndTimeSubscription: %w", err)
	}

	return s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
}

func (s *Suite) UserUpdatesCurrentLessonWithMissingField(ctx context.Context, missingField string) (context.Context, error) {
	return s.CommonSuite.UserUpdatesCurrentLessonWithMissingFields(ctx, missingField)
}

func (s *Suite) validateLessonForUpdateRequestMGMTInLessonmgmt(ctx context.Context, e *domain.Lesson, req *lpb.UpdateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if e.LessonID != req.LessonId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("Lesson.LessonID mismatched")
	}

	studentInfoList := make([]*lpb.CreateLessonRequest_StudentInfo, 0, len(req.StudentInfoList))
	for _, v := range req.StudentInfoList {
		studentInfoList = append(studentInfoList, &lpb.CreateLessonRequest_StudentInfo{
			StudentId:        v.StudentId,
			CourseId:         v.CourseId,
			AttendanceStatus: v.AttendanceStatus,
			LocationId:       v.LocationId,
			AttendanceNote:   v.AttendanceNote,
			AttendanceNotice: v.AttendanceNotice,
			AttendanceReason: v.AttendanceReason,
		})
	}

	return s.ValidateLessonForCreatedRequestMGMTInLessonmgmt(ctx, e, &lpb.CreateLessonRequest{
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		TeachingMedium:  req.TeachingMedium,
		TeachingMethod:  req.TeachingMethod,
		TeacherIds:      req.TeacherIds,
		LocationId:      req.LocationId,
		StudentInfoList: studentInfoList,
		Materials:       req.Materials,
		SavingOption: &lpb.CreateLessonRequest_SavingOption{
			Method: req.SavingOption.Method,
		},
		ClassId:          req.ClassId,
		CourseId:         req.CourseId,
		SchedulingStatus: req.SchedulingStatus,
	})
}
func (s *Suite) userUpdateBySavingTo(ctx context.Context, savingType, service string) (context.Context, error) {
	if service == "bob" {
		return s.userUpdateBySavingToInBob(ctx, savingType)
	}
	return s.userUpdateBySavingToInLessonmgmt(ctx, savingType)
}

func (s *Suite) userUpdateBySavingToInLessonmgmt(ctx context.Context, savingType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected *lpb.CreateLessonRequest, got %T", req)
	}
	lessonStatus := lpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
	teacherIds := []string{}
	studentInfo := []*lpb.UpdateLessonRequest_StudentInfo{}
	if savingType == "published" {
		lessonStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
		teacherIds = req.TeacherIds
		for _, studentInf := range req.StudentInfoList {
			studentInfo = append(studentInfo, &lpb.UpdateLessonRequest_StudentInfo{
				StudentId:        studentInf.StudentId,
				CourseId:         studentInf.CourseId,
				AttendanceStatus: studentInf.AttendanceStatus,
				AttendanceNote:   studentInf.AttendanceNote,
				AttendanceNotice: studentInf.AttendanceNotice,
				AttendanceReason: studentInf.AttendanceReason,
				LocationId:       req.LocationId,
			})
		}
	}

	updateLessonReq := &lpb.UpdateLessonRequest{
		LessonId:         stepState.CurrentLessonID,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		TeachingMedium:   req.TeachingMedium,
		TeachingMethod:   req.TeachingMethod,
		LocationId:       req.LocationId,
		SchedulingStatus: lessonStatus,
		TeacherIds:       teacherIds,
		StudentInfoList:  studentInfo,
		SavingOption: &lpb.UpdateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
	}
	return s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateLessonReq)
}

func (s *Suite) userUpdateBySavingToInBob(ctx context.Context, savingType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req, ok := stepState.Request.(*bpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected *bpb.CreateLessonRequest, got %T", req)
	}
	lessonStatus := bpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
	teacherIds := []string{}
	studentInfo := []*bpb.UpdateLessonRequest_StudentInfo{}
	if savingType == "published" {
		lessonStatus = bpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
		teacherIds = req.TeacherIds
		for _, studentInf := range req.StudentInfoList {
			studentInfo = append(studentInfo, &bpb.UpdateLessonRequest_StudentInfo{
				StudentId:        studentInf.StudentId,
				CourseId:         studentInf.CourseId,
				AttendanceStatus: studentInf.AttendanceStatus,
				LocationId:       studentInf.LocationId,
			})
		}
	}

	updateLessonReq := &bpb.UpdateLessonRequest{
		LessonId:         stepState.CurrentLessonID,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
		TeachingMedium:   req.TeachingMedium,
		TeachingMethod:   req.TeachingMethod,
		CenterId:         req.CenterId,
		SchedulingStatus: lessonStatus,
		TeacherIds:       teacherIds,
		StudentInfoList:  studentInfo,
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: bpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME,
		},
	}
	return s.CommonSuite.UserUpdateALessonWithRequest(ctx, updateLessonReq)
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

func (s *Suite) aDateInfoExistedInDB(ctx context.Context, date, locationID, dateType, openTime, status, resourcePath string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(dateType) < 1 || len(resourcePath) < 1 {
		return StepStateToContext(ctx, stepState), nil
	}

	dateTypestmt := "INSERT INTO day_type (day_type_id, resource_path) VALUES ($1, $2) ON CONFLICT DO NOTHING"
	if _, err := s.BobDB.Exec(ctx, dateTypestmt, dateType, resourcePath); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not init date type for this resource path: %w", err)
	}

	Date, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("parse datetime error: %w", err)
	}

	if openTime == NIL_VALUE {
		openTime = ""
	}
	dateInfostmt := `INSERT INTO day_info (date, location_id, day_type_id, opening_time, status, resource_path, time_zone)
					VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT DO NOTHING`
	if _, err := s.BobDB.Exec(ctx, dateInfostmt, Date, locationID, dateType, openTime, status, resourcePath, "Asia/Ho_Chi_Minh"); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not init date info: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) AnExistingLessonWithStudentAttendanceInfo(ctx context.Context, atttendanceStatus, atttendanceNotice, atttendanceReason, atttendanceNote string) (context.Context, error) {
	return s.UserCreateANewLessonWithStudentAttendanceInfoInLessonMgmt(ctx, atttendanceStatus, atttendanceNotice, atttendanceReason, atttendanceNote)
}

func (s *Suite) LocksLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonID := stepState.CurrentLessonID
	query := "UPDATE lessons l SET is_locked = true WHERE l.lesson_id = $1 "

	if _, err := s.LessonmgmtDB.Exec(ctx, query, lessonID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not update isLocked of lesson: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserUpdatesLessonStudentAttendanceInfoTo(ctx context.Context, atttendanceStatus, atttendanceNotice, atttendanceReason, atttendanceNote string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonID := stepState.CurrentLessonID
	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.CreateLessonRequest, got %T", createdRequest)
	}

	updateRequest := updateRequestFromCreateRequestInLessonmgmt(lessonID, createdRequest)

	for i := range updateRequest.StudentInfoList {
		updateRequest.StudentInfoList[i].AttendanceStatus = lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[atttendanceStatus])
		updateRequest.StudentInfoList[i].AttendanceNotice = lpb.StudentAttendanceNotice(lpb.StudentAttendanceNotice_value[atttendanceNotice])
		updateRequest.StudentInfoList[i].AttendanceReason = lpb.StudentAttendanceReason(lpb.StudentAttendanceReason_value[atttendanceReason])
		updateRequest.StudentInfoList[i].AttendanceNote = atttendanceNote
		updateRequest.StudentInfoList[i].LocationId = createdRequest.LocationId
	}

	return s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
}

func (s *Suite) TheAttendanceInfoIsUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	updateRequest, ok := stepState.Request.(*lpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.UpdateLessonRequest, got %T", updateRequest)
	}

	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	stepState.Lesson = lesson

	studentMap := make(map[string]*domain.LessonLearner)

	for _, st := range updateRequest.StudentInfoList {
		studentMap[st.StudentId] = &domain.LessonLearner{
			AttendStatus:     domain.StudentAttendStatus(st.AttendanceStatus.String()),
			AttendanceNote:   st.AttendanceNote,
			AttendanceNotice: domain.StudentAttendanceNotice(st.AttendanceNotice.String()),
			AttendanceReason: domain.StudentAttendanceReason(st.AttendanceReason.String()),
		}
	}

	for _, l := range lesson.Learners {
		if studentMap[l.LearnerID].AttendStatus != l.AttendStatus {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for `attendance_status`, got %s",
				string(studentMap[l.LearnerID].AttendStatus),
				string(l.AttendStatus))
		}

		if studentMap[l.LearnerID].AttendanceNote != l.AttendanceNote {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for `attendance_note`, got %s",
				studentMap[l.LearnerID].AttendanceNote,
				string(l.AttendanceNote))
		}

		if studentMap[l.LearnerID].AttendanceNotice != l.AttendanceNotice {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for `attendance_notice`, got %s",
				string(studentMap[l.LearnerID].AttendanceNotice),
				string(l.AttendanceNotice))
		}

		if studentMap[l.LearnerID].AttendanceReason != l.AttendanceReason {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for `attendance_reason`, got %s",
				string(studentMap[l.LearnerID].AttendanceReason),
				string(l.AttendanceReason))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheStudentAttendanceStatusIs(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonMemberRepo := repo.LessonMemberRepo{}
	member, err := lessonMemberRepo.FindByID(ctx, s.LessonmgmtDB, stepState.CurrentLessonID, stepState.StudentIDWithCourseID[0])
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to find lesson member by ID: %s", err)
	}
	if member.AttendanceStatus != status {
		return StepStateToContext(ctx, stepState), fmt.Errorf("student's attendance status is not updated: %s %s", member.LessonID, member.StudentID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) AnExistingLessonWithClassroom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.UserCreateANewLessonWithClassrooms(StepStateToContext(ctx, stepState), "existing")
}

func (s *Suite) UserUpdatesLessonClassroomWithRecord(ctx context.Context, recordState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonID := stepState.CurrentLessonID
	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.CreateLessonRequest, got %T", createdRequest)
	}

	updateRequest := updateRequestFromCreateRequestInLessonmgmt(lessonID, createdRequest)

	if recordState == "existing" {
		updateRequest.ClassroomIds = append(updateRequest.ClassroomIds, stepState.ClassroomIDs[0], stepState.ClassroomIDs[1])
	} else {
		updateRequest.ClassroomIds = append(updateRequest.ClassroomIds, stepState.ClassroomIDs[0], idutil.ULIDNow())
	}

	return s.CommonSuite.UserUpdateALessonWithRequestInLessonmgmt(ctx, updateRequest)
}

func (s *Suite) TheLessonClassroomsAreUpdated(ctx context.Context, updateState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.LessonmgmtDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	actualClassroomIDs := lesson.Classrooms.GetIDs()

	updatedRequest, ok := stepState.Request.(*lpb.UpdateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.UpdateLessonRequest, got %T", updatedRequest)
	}

	if updateState == "updated" && !stringutil.SliceElementsMatch(actualClassroomIDs, updatedRequest.ClassroomIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for updated lesson classroom IDs, got %s", updatedRequest.ClassroomIds, actualClassroomIDs)
	} else if updateState == "not updated" && stringutil.SliceElementsMatch(actualClassroomIDs, updatedRequest.ClassroomIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson classroom IDs are not updated, but got %s", actualClassroomIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) MarkStudentAsReallocate(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &lpb.MarkStudentAsReallocateRequest{
		StudentId: stepState.StudentIDWithCourseID[0],
		LessonId:  stepState.CurrentLessonID,
	}
	res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).MarkStudentAsReallocate(contextWithToken(s, ctx), req)
	stepState.ResponseErr = err
	stepState.Response = res
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}
