package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const timeLayout = "2006-01-02"

func (s *Suite) UserCreateLessonsWithSchedulingStatusAndTeachingMethod(ctx context.Context, status string, teachingMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch teachingMethod {
	case "individual":
		stepState.CurrentTeachingMethod = "individual"
	case "group":
		stepState.CurrentTeachingMethod = "group"
	}
	return s.CommonSuite.UserCreateSomeLessonsWithMissingFieldsAndSchedulingStatusInLessonmgmt(ctx, status)
}

func (s *Suite) UserCreateALessonWithAllRequiredFieldsInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"
	return s.CommonSuite.UserCreateALessonWithMissingFieldsInLessonmgmt(ctx)
}

func (s *Suite) UserCreateALessonWithTeachingMethodAndAllRequiredFieldsInLessonmgmt(ctx context.Context, teachingMethod string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = teachingMethod
	return s.CommonSuite.UserCreateALessonWithMissingFieldsInLessonmgmt(ctx)
}

func (s *Suite) UserCreateALessonWithAllRequiredFieldsWithSubInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.ValidateLessonCreatedSubscriptionInLessonmgmt(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.CreateLessonCreatedSubscription: %w", err)
	}

	return s.UserCreateALessonWithAllRequiredFieldsInLessonmgmt(ctx)
}

func (s *Suite) TheLessonWasCreatedInLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.LessonmgmtDB, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	stepState.Lesson = lesson

	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	if ctx, err = s.ValidateLessonForCreatedRequestMGMTInLessonmgmt(ctx, lesson, createdRequest); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for create Lesson: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ValidateLessonForCreatedRequestMGMTInLessonmgmt(ctx context.Context, e *domain.Lesson, req *lpb.CreateLessonRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if !e.StartTime.Equal(req.StartTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for start time, got %s", req.StartTime.AsTime(), e.StartTime)
	}
	if !e.EndTime.Equal(req.EndTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for end time, got %s", req.EndTime.AsTime(), e.EndTime)
	}

	if req.Materials != nil && len(req.GetMaterials()) > 0 {
		actualMediaIDs := make(map[string]bool)
		if e.Material != nil {
			for _, mediaID := range e.Material.MediaIDs {
				actualMediaIDs[mediaID] = true
			}
		}
		for _, expectedMediaID := range stepState.MediaIDs {
			if _, ok := actualMediaIDs[expectedMediaID]; !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("could not find media ID %s", expectedMediaID)
			}
		}
	}

	if e.LocationID != req.LocationId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected CenterId %s but got %s", req.LocationId, e.LocationID)
	}
	if req.TeachingMedium.String() != string(e.TeachingMedium) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMedium %s but got %s", req.TeachingMedium.String(), e.TeachingMedium)
	}
	if req.TeachingMethod.String() != string(e.TeachingMethod) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMethod %s but got %s", req.TeachingMethod.String(), e.TeachingMethod)
	}
	if e.SchedulingStatus != domain.LessonSchedulingStatus(req.SchedulingStatus.String()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected SchedulingStatus %s but got %s", domain.LessonSchedulingStatus(req.SchedulingStatus.String()), e.SchedulingStatus)
	}
	zoomInfo := req.GetZoomInfo()
	if zoomInfo != nil {
		if zoomInfo.GetZoomLink() != string(e.ZoomLink) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected ZoomLink %s but got %s", zoomInfo.GetZoomLink(), e.ZoomLink)
		}
		if zoomInfo.GetZoomAccountOwner() != e.ZoomOwnerID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected ZoomAccountOwner %s but got %s", zoomInfo.GetZoomAccountOwner(), e.ZoomOwnerID)
		}
	}
	classDoInfo := req.GetClassDoInfo()
	if classDoInfo != nil {
		if classDoInfo.GetClassDoOwnerId() != e.ClassDoOwnerID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected ZoomAccountOwner %s but got %s", classDoInfo.GetClassDoOwnerId(), e.ClassDoOwnerID)
		}
		if classDoInfo.GetClassDoLink() != e.ClassDoLink {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected ZoomLink %s but got %s", classDoInfo.GetClassDoLink(), e.ClassDoLink)
		}
		if classDoInfo.GetClassDoRoomId() != e.ClassDoRoomID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected ZoomLink %s but got %s", classDoInfo.GetClassDoRoomId(), e.ClassDoRoomID)
		}
	}

	actualTeacherIDs := e.GetTeacherIDs()
	if !stringutil.SliceElementsMatch(actualTeacherIDs, req.TeacherIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher IDs, got %s", req.TeacherIds, actualTeacherIDs)
	}

	learnerIds := make([]string, 0, len(req.StudentInfoList))
	for _, studentInfo := range req.StudentInfoList {
		learnerIds = append(learnerIds, studentInfo.StudentId)
	}
	actualLearnerIDs := e.GetLearnersIDs()
	if !stringutil.SliceElementsMatch(actualLearnerIDs, learnerIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for learner IDs, got %s", learnerIds, actualLearnerIDs)
	}
	// TODO: check course and location of lesson members
	// Validate lesson group
	if req.TeachingMethod == cpb.LessonTeachingMethod_LESSON_TEACHING_METHOD_GROUP {
		if req.CourseId != string(e.CourseID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected CourseID %s but got %s", req.CourseId, e.CourseID)
		}
		if req.ClassId != string(e.ClassID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected ClassID %s but got %s", req.ClassId, e.ClassID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateLesson(ctx context.Context, lessonStatus, mFields, service string) (context.Context, error) {
	if service == "bob" {
		return s.userCreateLessonInBob(ctx, lessonStatus, mFields)
	}
	return s.userCreateLessonInLessonmgmt(ctx, lessonStatus, mFields)
}

func (s *Suite) userCreateLessonInBob(ctx context.Context, lessonStatus, mFields string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var missingFields []string
	schedulingStatus := bpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
	switch mFields {
	case "students":
		missingFields = append(missingFields, "students")
	case "teachers":
		missingFields = append(missingFields, "teachers")
	case "students and teachers":
		missingFields = append(missingFields, "students", "teachers")
	}
	if lessonStatus == "draft" {
		schedulingStatus = bpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
	}
	req := s.CommonSuite.UserCreateALessonRequestWithMissingFields(StepStateToContext(ctx, stepState), cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE, missingFields...)
	req.SchedulingStatus = schedulingStatus
	return s.CommonSuite.UserCreateALessonWithRequest(ctx, req)
}

func (s *Suite) userCreateLessonInLessonmgmt(ctx context.Context, lessonStatus, mFields string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var missingFields []string
	schedulingStatus := lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
	switch mFields {
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
	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(StepStateToContext(ctx, stepState), cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE, missingFields...)
	req.SchedulingStatus = schedulingStatus
	req.LessonCapacity = 20
	return s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *Suite) UserCreateANewLessonWithDateLocationAndOtherRequiredFieldsInLessonmgmt(ctx context.Context, date, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	convertedDate, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return ctx, fmt.Errorf("parse datetime error: %w", err)
	}

	stepState.CurrentTeachingMethod = "individual"

	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE)
	indexLocation, _ := strconv.Atoi(location)
	locationID := stepState.CenterIDs[indexLocation]
	req.LocationId = locationID
	for i := range req.StudentInfoList {
		req.StudentInfoList[i].LocationId = locationID
	}
	req.StartTime = timestamppb.New(convertedDate)
	req.EndTime = timestamppb.New(convertedDate.Add(2 * time.Hour))

	return s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *Suite) TheLessonWasInLessonmgmt(ctx context.Context, createStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if createStatus == "created" {
		return s.TheLessonWasCreatedInLessonmgmt(ctx)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserCreateANewLessonWithStudentAttendanceInfoInLessonMgmt(ctx context.Context, atttendanceStatus, atttendanceNotice, atttendanceReason, atttendanceNote string) (context.Context, error) {
	var studentInfoList []*lpb.CreateLessonRequest_StudentInfo

	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"

	studentID := stepState.StudentIDWithCourseID[0]
	courseID := stepState.StudentIDWithCourseID[1]
	locationID := stepState.CenterIDs[len(stepState.CenterIDs)-1]
	studentInfoList = append(studentInfoList, &lpb.CreateLessonRequest_StudentInfo{
		StudentId:        studentID,
		CourseId:         courseID,
		AttendanceStatus: lpb.StudentAttendStatus(lpb.StudentAttendStatus_value[atttendanceStatus]),
		AttendanceNotice: lpb.StudentAttendanceNotice(lpb.StudentAttendanceNotice_value[atttendanceNotice]),
		AttendanceReason: lpb.StudentAttendanceReason(lpb.StudentAttendanceReason_value[atttendanceReason]),
		AttendanceNote:   atttendanceNote,
		LocationId:       locationID,
	})

	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE)
	req.StudentInfoList = studentInfoList

	return s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *Suite) TheAttendanceInfoIsCorrect(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.CreateLessonRequest, got %T", createdRequest)
	}

	lesson := stepState.Lesson
	if err := s.isCorrectLessonLearner(ctx, lesson.Learners, createdRequest, true); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserCreateANewLessonWithClassrooms(ctx context.Context, recordState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"
	classroomIDs := []string{}

	if recordState == "existing" {
		classroomIDs = append(classroomIDs, stepState.ClassroomIDs[len(stepState.ClassroomIDs)-1], stepState.ClassroomIDs[len(stepState.ClassroomIDs)-2])
	} else {
		classroomIDs = append(classroomIDs, stepState.ClassroomIDs[len(stepState.ClassroomIDs)-1], idutil.ULIDNow())
	}

	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, cpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE)
	req.ClassroomIds = classroomIDs

	return s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *Suite) TheClassroomsAreInTheLesson(ctx context.Context, recordState string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if recordState == "existing" {
		lesson := stepState.Lesson
		createdRequest, ok := stepState.Request.(*lpb.CreateLessonRequest)
		if !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *lpb.CreateLessonRequest, got %T", createdRequest)
		}

		actualClassroomIDs := lesson.Classrooms.GetIDs()
		if !stringutil.SliceElementsMatch(actualClassroomIDs, createdRequest.ClassroomIds) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for classroom IDs, got %s", createdRequest.ClassroomIds, actualClassroomIDs)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
