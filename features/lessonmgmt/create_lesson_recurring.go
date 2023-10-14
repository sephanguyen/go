package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) createRecurringLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	loc := LoadLocalLocation()
	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(time.Date(2022, 7, 9, 9, 0, 0, 0, loc)),
		EndTime:         timestamppb.New(time.Date(2022, 7, 9, 10, 0, 0, 0, loc)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      stepState.CenterIDs[len(stepState.CenterIDs)-1],
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
			LocationId:       stepState.CenterIDs[len(stepState.CenterIDs)-1],
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

func (s *Suite) createRecurringLessonWithZoomInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	loc := LoadLocalLocation()
	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(time.Date(2022, 7, 9, 9, 0, 0, 0, loc)),
		EndTime:         timestamppb.New(time.Date(2022, 7, 9, 10, 0, 0, 0, loc)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_ZOOM,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      stepState.CenterIDs[len(stepState.CenterIDs)-1],
		StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*lpb.Material{},
		SavingOption: &lpb.CreateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
			Recurrence: &lpb.Recurrence{
				EndDate: timestamppb.New(time.Date(2022, 7, 31, 10, 0, 0, 0, loc)),
			},
		},
		ZoomInfo: &lpb.ZoomInfo{
			ZoomLink:         "htpp://zoom-link",
			ZoomAccountOwner: stepState.ZoomAccount.ID,
			ZoomId:           "123",
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
	query := `select lesson_id from lessons 
	where scheduler_id = (select scheduler_id from lessons 
						  where lesson_id = $1) order by start_time asc`
	rows, err := s.BobDB.Query(ctx, query, stepState.CurrentLessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lessons := []*domain.Lesson{}
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

func (s *Suite) equalTime(ctx context.Context, lesson *domain.Lesson, startTime, endTime time.Time) error {
	if !lesson.StartTime.Equal(startTime) {
		return fmt.Errorf("expected %s for start time, got %s", endTime, lesson.StartTime)
	}
	if !lesson.EndTime.Equal(endTime) {
		return fmt.Errorf("expected %s for end time, got %s", endTime, lesson.EndTime)
	}
	return nil
}

func (s *Suite) equalLesson(ctx context.Context, lesson *domain.Lesson, req *lpb.CreateLessonRequest) error {
	if lesson.LocationID != req.LocationId {
		return fmt.Errorf("expected CenterId %s but got %s", req.LocationId, lesson.LocationID)
	}
	if req.TeachingMedium.String() != string(lesson.TeachingMedium) {
		return fmt.Errorf("expected TeachingMedium %s but got %s", req.TeachingMedium.String(), lesson.TeachingMedium)
	}
	if req.TeachingMethod.String() != string(lesson.TeachingMethod) {
		return fmt.Errorf("expected TeachingMethod %s but got %s", req.TeachingMethod.String(), lesson.TeachingMethod)
	}
	if lesson.SchedulingStatus != domain.LessonSchedulingStatus(req.SchedulingStatus.String()) {
		return fmt.Errorf("expected SchedulingStatus %s but got %s", domain.LessonSchedulingStatus(req.SchedulingStatus.String()), lesson.SchedulingStatus)
	}

	actualTeacherIDs := lesson.GetTeacherIDs()
	if !stringutil.SliceElementsMatch(actualTeacherIDs, req.TeacherIds) {
		return fmt.Errorf("expected %s for teacher IDs, got %s", req.TeacherIds, actualTeacherIDs)
	}

	if req.TeachingMethod == cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP {
		if req.CourseId != string(lesson.CourseID) {
			return fmt.Errorf("expected CourseID %s but got %s", req.CourseId, lesson.CourseID)
		}
		if req.ClassId != string(lesson.ClassID) {
			return fmt.Errorf("expected ClassID %s but got %s", req.ClassId, lesson.ClassID)
		}
	}
	return nil
}

func (s *Suite) equalMaterial(ctx context.Context, lesson *domain.Lesson, baseLesson bool) error {
	stepState := StepStateFromContext(ctx)
	actualMediaIDs := make(map[string]bool)
	if lesson.Material != nil {
		for _, mediaID := range lesson.Material.MediaIDs {
			actualMediaIDs[mediaID] = true
		}
	}
	if baseLesson {
		for _, expectedMediaID := range stepState.MediaIDs {
			if _, ok := actualMediaIDs[expectedMediaID]; !ok {
				return fmt.Errorf("could not find media ID %s", expectedMediaID)
			}
		}
	} else if len(actualMediaIDs) > 0 {
		return fmt.Errorf("expected no material for %s lesson", lesson.LessonID)
	}
	return nil
}

func (s *Suite) isCorrectLessonLearner(ctx context.Context, learners []*domain.LessonLearner, req *lpb.CreateLessonRequest, baseLesson bool) error {
	studentMap := make(map[string]*domain.LessonLearner)
	for _, st := range req.StudentInfoList {
		studentMap[st.StudentId] = &domain.LessonLearner{
			AttendStatus:     domain.StudentAttendStatusEmpty,
			AttendanceNote:   "",
			AttendanceNotice: domain.NoticeEmpty,
			AttendanceReason: domain.ReasonEmpty,
		}
		if baseLesson {
			studentMap[st.StudentId].AttendStatus = domain.StudentAttendStatus(st.AttendanceStatus.String())
			studentMap[st.StudentId].AttendanceNote = st.AttendanceNote
			studentMap[st.StudentId].AttendanceNotice = domain.StudentAttendanceNotice(st.AttendanceNotice.String())
			studentMap[st.StudentId].AttendanceReason = domain.StudentAttendanceReason(st.AttendanceReason.String())
		}
	}
	for _, l := range learners {
		if studentMap[l.LearnerID].AttendStatus != l.AttendStatus {
			return fmt.Errorf("expected %s for `attendance_status`, got %s",
				string(studentMap[l.LearnerID].AttendStatus),
				string(l.AttendStatus))
		}

		if studentMap[l.LearnerID].AttendanceNote != l.AttendanceNote {
			return fmt.Errorf("expected %s for `attendance_note`, got %s",
				studentMap[l.LearnerID].AttendanceNote,
				string(l.AttendanceNote))
		}

		if studentMap[l.LearnerID].AttendanceNotice != l.AttendanceNotice {
			return fmt.Errorf("expected %s for `attendance_notice`, got %s",
				string(studentMap[l.LearnerID].AttendanceNotice),
				string(l.AttendanceNotice))
		}

		if studentMap[l.LearnerID].AttendanceReason != l.AttendanceReason {
			return fmt.Errorf("expected %s for `attendance_reason`, got %s",
				string(studentMap[l.LearnerID].AttendanceReason),
				string(l.AttendanceReason))
		}
	}
	return nil
}

func (s *Suite) hasRecurringLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*lpb.CreateLessonRequest)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", stepState.ResponseErr)
	}
	lessons, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	startTime := golibs.TimestamppbToTime(req.StartTime)
	endTime := golibs.TimestamppbToTime(req.EndTime)
	baseLesson := true
	studentCourseEndDate := stepState.EndDate
	for _, lesson := range lessons {
		// check lesson
		if err := s.equalLesson(ctx, lesson, req); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		// check time lesson
		if err := s.equalTime(ctx, lesson, startTime, endTime); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		if startTime.After(studentCourseEndDate) {
			if len(lesson.Learners) > 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson `%s` have student info not correct", lesson.LessonID)
			}
		}
		startTime = startTime.AddDate(0, 0, 7)
		endTime = endTime.AddDate(0, 0, 7)
		if err := s.isCorrectLessonLearner(ctx, lesson.Learners, req, baseLesson); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
		if req.Materials != nil && len(req.GetMaterials()) > 0 {
			if err := s.equalMaterial(ctx, lesson, baseLesson); err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
			}
		}
		baseLesson = false
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createRecurringLessonWithMissingFields(ctx context.Context, lessonStatus, mFields string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now().Round(time.Second)
	var missingFields []string
	schedulingStatus := lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
	switch mFields {
	case "none":
	case "students":
		missingFields = append(missingFields, "students")
	case "teachers":
		missingFields = append(missingFields, "teachers")
	case "students and teachers":
		missingFields = append(missingFields, "students", "teachers")
	}
	if lessonStatus == "draft" {
		schedulingStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
	}
	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE, missingFields...)
	req.SchedulingStatus = schedulingStatus
	req.SavingOption = &lpb.CreateLessonRequest_SavingOption{
		Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
		Recurrence: &lpb.Recurrence{
			EndDate: timestamppb.New(now.AddDate(0, 1, 0)),
		},
	}
	res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).CreateLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req
	stepState.ResponseErr = err
	stepState.Response = res
	stepState.CurrentLessonID = res.GetId()
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createRecurringLessonWithDateAndLocationUntilEndDate(ctx context.Context, date, location, endDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	convertedDate, err := time.Parse(timeLayout, date)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for lesson date: %w", err)
	}

	convertedEndDate, err := time.Parse(timeLayout, endDate)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error for lesson end date: %w", err)
	}
	indexLocation, _ := strconv.Atoi(location)
	locationID := stepState.CenterIDs[indexLocation]
	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(convertedDate),
		EndTime:         timestamppb.New(convertedDate.Add(2 * time.Hour)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      locationID,
		StudentInfoList: []*lpb.CreateLessonRequest_StudentInfo{},
		Materials:       []*lpb.Material{},
		SavingOption: &lpb.CreateLessonRequest_SavingOption{
			Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE,
			Recurrence: &lpb.Recurrence{
				EndDate: timestamppb.New(convertedEndDate),
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
			LocationId:       locationID,
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
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) recurringLessonsWillInclude(ctx context.Context, expectedDates, skippedDates string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonDates := stepState.LessonDates
	if len(expectedDates) > 0 {
		expectedDatesList := strings.Split(expectedDates, ",")

		for _, expectedDate := range expectedDatesList {
			if _, present := lessonDates[expectedDate]; !present {
				return ctx, fmt.Errorf("expected date %s not present in lesson dates", expectedDate)
			}
		}
		if len(expectedDatesList) != len(lessonDates) {
			return ctx, fmt.Errorf("invalid lesson chain, expect %s: actual %s", expectedDatesList, lessonDates)
		}
	}

	if len(skippedDates) > 0 {
		skippedDatesList := strings.Split(skippedDates, ",")

		for _, skippedDate := range skippedDatesList {
			if _, present := lessonDates[skippedDate]; present {
				return ctx, fmt.Errorf("skipped date %s present in lesson dates", skippedDate)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) CreateRecurringLessonWithClassrooms(ctx context.Context, recordState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	loc := LoadLocalLocation()

	req := &lpb.CreateLessonRequest{
		StartTime:       timestamppb.New(time.Date(2022, 7, 9, 9, 0, 0, 0, loc)),
		EndTime:         timestamppb.New(time.Date(2022, 7, 9, 10, 0, 0, 0, loc)),
		TeachingMedium:  cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE,
		TeachingMethod:  cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_INDIVIDUAL,
		TeacherIds:      stepState.TeacherIDs,
		LocationId:      stepState.CenterIDs[len(stepState.CenterIDs)-1],
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
			LocationId:       stepState.CenterIDs[len(stepState.CenterIDs)-1],
		})
	}

	for _, mediaID := range stepState.MediaIDs {
		req.Materials = append(req.Materials, &lpb.Material{
			Resource: &lpb.Material_MediaId{
				MediaId: mediaID,
			},
		})
	}

	if recordState == "existing" {
		req.ClassroomIds = append(req.ClassroomIds, stepState.ClassroomIDs[len(stepState.ClassroomIDs)-1], stepState.ClassroomIDs[len(stepState.ClassroomIDs)-2])
	} else {
		req.ClassroomIds = append(req.ClassroomIds, stepState.ClassroomIDs[len(stepState.ClassroomIDs)-1], idutil.ULIDNow())
	}

	stepState.Request = req
	res, err := lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).CreateLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.ResponseErr = err

	if recordState == "existing" {
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create recurring lesson: %w", err)
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
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theRecurringLessonIs(ctx context.Context, createStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if createStatus == "created" {
		return s.hasRecurringLesson(StepStateToContext(ctx, stepState))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) TheClassroomsAreInTheRecurringLesson(ctx context.Context, recordState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if recordState == "existing" {
		req := stepState.Request.(*lpb.CreateLessonRequest)
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", stepState.ResponseErr)
		}

		lessons, err := s.retrieveLessonChainByLessonID(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}

		for _, lesson := range lessons {
			actualClassroomIDs := lesson.Classrooms.GetIDs()
			if !stringutil.SliceElementsMatch(actualClassroomIDs, req.ClassroomIds) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for classroom IDs, got %s", req.ClassroomIds, actualClassroomIDs)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
