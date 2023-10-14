package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/stringutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	commonpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *suite) anExistingLesson(ctx context.Context, teachingMethod string) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = teachingMethod

	req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(ctx, commonpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE)
	req.ClassroomIds = append(req.ClassroomIds, stepState.ClassroomIDs...)

	return s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
}

func (s *suite) getLessonDetailOnCalendar(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(10 * time.Second)

	req := &cpb.GetLessonDetailOnCalendarRequest{
		LessonId: stepState.CurrentLessonID,
	}
	ctx = s.signedCtx(StepStateToContext(ctx, stepState))
	stepState.Response, stepState.ResponseErr = cpb.NewLessonReaderServiceClient(s.CalendarConn).
		GetLessonDetailOnCalendar(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) lessonDetailMatchesLessonCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createdLesson := stepState.Request.(*lpb.CreateLessonRequest)
	receivedLesson := stepState.Response.(*cpb.GetLessonDetailOnCalendarResponse)

	if receivedLesson == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected return lesson, got nil")
	}

	if !createdLesson.StartTime.AsTime().Equal(receivedLesson.StartTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for start time, got %s",
			createdLesson.StartTime.AsTime(),
			receivedLesson.StartTime.AsTime())
	}
	if !createdLesson.EndTime.AsTime().Equal(receivedLesson.EndTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for end time, got %s",
			createdLesson.EndTime.AsTime(),
			receivedLesson.EndTime.AsTime())
	}

	if createdLesson.TeachingMedium.String() != receivedLesson.TeachingMedium.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected teaching medium %s but got %s",
			createdLesson.TeachingMedium.String(),
			receivedLesson.TeachingMedium.String())
	}
	if createdLesson.TeachingMethod.String() != receivedLesson.TeachingMethod.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected teaching method %s but got %s",
			createdLesson.TeachingMethod.String(),
			receivedLesson.TeachingMethod.String())
	}
	if createdLesson.SchedulingStatus.String() != receivedLesson.SchedulingStatus.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected scheduling status %s but got %s",
			createdLesson.SchedulingStatus.String(),
			receivedLesson.SchedulingStatus.String())
	}

	if createdLesson.LocationId != receivedLesson.Location.LocationId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected location id %s but got %s",
			createdLesson.LocationId,
			receivedLesson.Location.LocationId)
	}

	actualTeacherIDs := make([]string, 0, len(receivedLesson.LessonTeachers))
	for _, teacher := range receivedLesson.LessonTeachers {
		actualTeacherIDs = append(actualTeacherIDs, teacher.TeacherId)
	}
	if !stringutil.SliceElementsMatch(createdLesson.TeacherIds, actualTeacherIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher IDs but got %s",
			createdLesson.TeacherIds,
			actualTeacherIDs)
	}

	learnerIDs := make([]string, 0, len(createdLesson.StudentInfoList))
	for _, student := range createdLesson.StudentInfoList {
		learnerIDs = append(learnerIDs, student.StudentId)
	}
	actualLearnerIDs := make([]string, 0, len(receivedLesson.LessonMembers))
	for _, lessonMember := range receivedLesson.LessonMembers {
		actualLearnerIDs = append(actualLearnerIDs, lessonMember.StudentId)
	}
	if !stringutil.SliceElementsMatch(learnerIDs, actualLearnerIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for learner IDs but got %s",
			learnerIDs,
			actualLearnerIDs)
	}

	actualClassroomIDs := make([]string, 0, len(receivedLesson.LessonClassrooms))
	for _, classroom := range receivedLesson.LessonClassrooms {
		actualClassroomIDs = append(actualClassroomIDs, classroom.ClassroomId)
	}
	if !stringutil.SliceElementsMatch(createdLesson.ClassroomIds, actualClassroomIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for classroom IDs but got %s",
			createdLesson.ClassroomIds,
			actualClassroomIDs)
	}

	actualMediaIDs := make(map[string]bool)
	for _, mediaID := range receivedLesson.MediaIds {
		actualMediaIDs[mediaID] = true
	}
	for _, expectedMediaID := range stepState.MediaIDs {
		if _, ok := actualMediaIDs[expectedMediaID]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("could not find media ID %s", expectedMediaID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
