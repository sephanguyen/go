package lessonmgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func updateRequestFromCreateRequest(lessonID string, createdRequest *bpb.CreateLessonRequest) *bpb.UpdateLessonRequest {
	updateRequest := &bpb.UpdateLessonRequest{
		LessonId:        lessonID,
		StartTime:       createdRequest.StartTime,
		EndTime:         createdRequest.EndTime,
		TeachingMedium:  createdRequest.TeachingMedium,
		TeachingMethod:  createdRequest.TeachingMethod,
		TeacherIds:      createdRequest.TeacherIds,
		CenterId:        createdRequest.CenterId,
		StudentInfoList: make([]*bpb.UpdateLessonRequest_StudentInfo, 0, len(createdRequest.StudentInfoList)),
		Materials:       createdRequest.Materials,
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: createdRequest.SavingOption.Method,
		},
	}
	for _, studentInf := range createdRequest.StudentInfoList {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentInf.StudentId,
			CourseId:         studentInf.CourseId,
			AttendanceStatus: studentInf.AttendanceStatus,
			LocationId:       createdRequest.CenterId,
		})
	}

	return updateRequest
}

func (s *Suite) updateLessonCenterID(ctx context.Context) (string, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CenterIDs = []string{}
	stepState.LocationIDs = []string{}
	ctx, err := s.someCenters(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create center: %s", err)
	}

	return stepState.CenterIDs[len(stepState.CenterIDs)-1], nil
}

func (s *Suite) updateLessonTeacherIDs(ctx context.Context) ([]string, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDs = []string{}
	_, err := s.CreateTeacherAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create teachers: %s", err)
	}
	if len(stepState.TeacherIDs) < 1 {
		return nil, fmt.Errorf("required at least 1 teacher, got %d", len(stepState.TeacherIDs))
	}

	return stepState.TeacherIDs, nil
}

func (s *Suite) updateStudentInfo(ctx context.Context) ([]*bpb.UpdateLessonRequest_StudentInfo, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentIds = []string{}
	ctx, err := s.CommonSuite.CreateStudentAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create students: %s", err)
	}
	if len(stepState.StudentIds) < 1 {
		return nil, fmt.Errorf("required at least 1 student, got %d", len(stepState.StudentIds))
	}
	ctx, err = s.CommonSuite.SomeStudentSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create student subscription: %s", err)
	}

	studentInfoList := make([]*bpb.UpdateLessonRequest_StudentInfo, 0, len(stepState.StudentIds))
	for _, studentID := range stepState.StudentIds {
		studentInfoList = append(studentInfoList,
			&bpb.UpdateLessonRequest_StudentInfo{
				StudentId:        studentID,
				CourseId:         stepState.CourseIDs[len(stepState.CourseIDs)-1],
				AttendanceStatus: bpb.StudentAttendStatus_STUDENT_ATTEND_STATUS_INFORMED_ABSENT,
				LocationId:       stepState.CenterIDs[len(stepState.CenterIDs)-1],
			},
		)
	}

	return studentInfoList, nil
}

func (s *Suite) updateMaterial(ctx context.Context) ([]*bpb.Material, error) {
	stepState := StepStateFromContext(ctx)
	stepState.MediaIDs = []string{}
	ctx, err := s.CommonSuite.UpsertValidMediaList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create media: %s", err)
	}
	if len(stepState.MediaIDs) < 1 {
		return nil, fmt.Errorf("required at least 1 media, got %d", len(stepState.MediaIDs))
	}

	materials := make([]*bpb.Material, 0, len(stepState.MediaIDs))
	for _, id := range stepState.MediaIDs {
		materials = append(materials,
			&bpb.Material{
				Resource: &bpb.Material_MediaId{
					MediaId: id,
				},
			},
		)
	}
	return materials, nil
}

//nolint:gocyclo
func (s *Suite) userUpdatesFieldInTheLesson(ctx context.Context, fieldName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	randID := idutil.ULIDNow()
	createdRequest, ok := stepState.Request.(*bpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	updateRequest := updateRequestFromCreateRequest(stepState.CurrentLessonID, createdRequest)

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
		updateRequest.CenterId, err = s.updateLessonCenterID(s.returnRootContext(ctx))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		for i := range updateRequest.StudentInfoList {
			updateRequest.StudentInfoList[i].LocationId = updateRequest.CenterId
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
		studentInfoList, err := s.updateStudentInfo(s.returnRootContext(ctx))
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
		materials, err := s.updateMaterial(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		updateRequest.Materials = append(updateRequest.Materials, materials...)
	}
	stepState.CurrentLessonID = updateRequest.LessonId
	stepState.Request = updateRequest
	ctx, err = s.CommonSuite.CreateEditLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditFieldLessonSubscription: %w", err)
	}

	return s.CommonSuite.UserUpdateALessonWithRequest(ctx, updateRequest)
}

func (s *Suite) createClassAndCourseWithID(ctx context.Context, classID string, courseID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	className := "bdd-test-class-name"
	courseName := "bdd-test-course-name"
	locationID := constants.ManabieOrgLocation
	schoolID := golibs.ResourcePathFromCtx(ctx)
	classFields := []string{"class_id", "name", "course_id", "location_id", "school_id", "created_at", "updated_at"}
	courseFields := []string{"course_id", "name", "school_id", "created_at", "updated_at"}
	insertClassQuery := fmt.Sprintf("INSERT INTO class (%s) VALUES ($1,$2,$3,$4,$5,$6,$7)",
		strings.Join(classFields, ","))
	insertCourseQuery := fmt.Sprintf("INSERT INTO courses (%s) VALUES ($1,$2,$3,$4,$5)",
		strings.Join(courseFields, ","))
	now := time.Now()
	stepState.RequestSentAt = now
	_, err := s.BobDBTrace.Exec(ctx, insertCourseQuery, courseID, courseName, schoolID, now, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	_, err = s.BobDBTrace.Exec(ctx, insertClassQuery, classID, className, courseID, locationID, schoolID, now, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdatesStatusInTheLessonIsValue(ctx context.Context, value string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := &lpb.UpdateLessonSchedulingStatusRequest{
		LessonId: stepState.CurrentLessonID,
	}
	switch value {
	case cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT.String():
		request.SchedulingStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT
	case cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String():
		request.SchedulingStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED
	case cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String():
		request.SchedulingStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
	case cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String():
		request.SchedulingStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("LessonSchedulingStatus is not available")
	}

	stepState.CurrentLessonID = request.LessonId
	stepState.Request = request

	ctx, err := s.CommonSuite.CreateEditStatusLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditFieldLessonSubscription: %w", err)
	}
	return s.CommonSuite.UserUpdateALessonSchedulingStatusWithRequest(ctx, request)
}

func (s *Suite) TheLessonSchedulingStatusWasUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.SavingType == lpb.SavingType_THIS_AND_FOLLOWING {
		return StepStateToContext(ctx, stepState), s.checkSchedulingStatusByThisAndFollowing(ctx)
	} else {
		return StepStateToContext(ctx, stepState), s.checkSchedulingStatusByOnlyThis(ctx)
	}
}

func (s *Suite) checkSchedulingStatusByOnlyThis(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	lesson, err := lessonRepo.GetLessonByID(ctx, s.CommonSuite.BobDB, stepState.CurrentLessonID)
	if err != nil {
		return fmt.Errorf("failed to query lesson: %s", err)
	}
	updatedRequest, ok := stepState.Request.(*lpb.UpdateLessonSchedulingStatusRequest)
	if !ok {
		return fmt.Errorf("expected stepState.Request to be *lpb.UpdateLessonSchedulingStatusRequest, got %T", updatedRequest)
	}
	if lesson.SchedulingStatus != domain.LessonSchedulingStatus(updatedRequest.SchedulingStatus.String()) {
		return fmt.Errorf("expected SchedulingStatus %s but got %s", domain.LessonSchedulingStatus(updatedRequest.SchedulingStatus.String()), lesson.SchedulingStatus)
	}
	return nil
}

func (s *Suite) checkSchedulingStatusByThisAndFollowing(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	lessonChain, err := s.retrieveLessonChainByLessonID(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve lesson chain :%w", err)
	}
	updatedRequest, ok := stepState.Request.(*lpb.UpdateLessonSchedulingStatusRequest)
	if !ok {
		return fmt.Errorf("expected stepState.Request to be *lpb.UpdateLessonSchedulingStatusRequest, got %T", updatedRequest)
	}
	for _, ls := range lessonChain {
		if ls.SchedulingStatus != domain.LessonSchedulingStatus(updatedRequest.SchedulingStatus.String()) {
			return fmt.Errorf("expected %s,but got %s", domain.LessonSchedulingStatus(updatedRequest.SchedulingStatus.String()), ls.SchedulingStatus)
		}
	}
	return nil
}

func (s *Suite) userUpdateALessonWithStartTimeLaterThanEndTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createdRequest, ok := stepState.Request.(*bpb.CreateLessonRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.UpdateLiveLessonRequest, got %T", createdRequest)
	}
	updateRequest := &bpb.UpdateLessonRequest{
		LessonId:        stepState.CurrentLessonID,
		StartTime:       createdRequest.StartTime,
		EndTime:         createdRequest.EndTime,
		TeachingMedium:  createdRequest.TeachingMedium,
		TeachingMethod:  createdRequest.TeachingMethod,
		TeacherIds:      createdRequest.TeacherIds,
		CenterId:        createdRequest.CenterId,
		StudentInfoList: make([]*bpb.UpdateLessonRequest_StudentInfo, 0, len(createdRequest.StudentInfoList)),
		Materials:       createdRequest.Materials,
		SavingOption: &bpb.UpdateLessonRequest_SavingOption{
			Method: createdRequest.SavingOption.Method,
		},
	}
	for _, studentInf := range createdRequest.StudentInfoList {
		updateRequest.StudentInfoList = append(updateRequest.StudentInfoList, &bpb.UpdateLessonRequest_StudentInfo{
			StudentId:        studentInf.StudentId,
			CourseId:         studentInf.CourseId,
			AttendanceStatus: studentInf.AttendanceStatus,
		})
	}

	updateRequest.StartTime = timestamppb.New(updateRequest.StartTime.AsTime().Add(2 * time.Hour))
	updateRequest.EndTime = timestamppb.New(updateRequest.StartTime.AsTime().Add(-time.Hour))
	stepState.Request = updateRequest
	ctx, err := s.CommonSuite.CreateEditLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEditLessonWithStartTimeLaterThanEndTimeSubscription: %w", err)
	}

	return s.CommonSuite.UserUpdateALessonWithRequest(ctx, updateRequest)
}

func (s *Suite) UserUpdatesCurrentLessonWithMissingFieldInLessonmgmt(ctx context.Context, missingField string) (context.Context, error) {
	return s.CommonSuite.UserUpdatesCurrentLessonWithMissingFieldsInLessonmgmt(ctx, missingField)
}
