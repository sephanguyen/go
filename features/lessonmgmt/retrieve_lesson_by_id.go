package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) RetrieveLessonByID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &lpb.RetrieveLessonByIDRequest{
		LessonId: stepState.CurrentLessonID,
	}
	stepState.Response, stepState.ResponseErr = lpb.NewLessonReaderServiceClient(s.Connections.LessonMgmtConn).
		RetrieveLessonByID(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) LessonMatchWithLessonCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lesson := stepState.Response.(*lpb.RetrieveLessonByIDResponse).Lesson
	if lesson == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected return lesson, got nil")
	}
	req := stepState.Request.(*bpb.CreateLessonRequest)
	if !lesson.StartTime.AsTime().Equal(req.StartTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for start time, got %s", req.StartTime.AsTime(), lesson.StartTime)
	}
	if !lesson.EndTime.AsTime().Equal(req.EndTime.AsTime()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for end time, got %s", req.EndTime.AsTime(), lesson.EndTime)
	}

	actualMediaIDs := make(map[string]bool)
	for _, mediaID := range lesson.MediaIds {
		actualMediaIDs[mediaID] = true
	}
	for _, expectedMediaID := range stepState.MediaIDs {
		if _, ok := actualMediaIDs[expectedMediaID]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("could not find media ID %s", expectedMediaID)
		}
	}

	if lesson.LocationId != req.CenterId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected CenterId %s but got %s", req.CenterId, lesson.LocationId)
	}
	if req.TeachingMedium.String() != lesson.TeachingMedium.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMedium %s but got %s", req.TeachingMedium.String(), lesson.TeachingMedium)
	}
	if req.TeachingMethod.String() != lesson.TeachingMethod.String() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected TeachingMethod %s but got %s", req.TeachingMethod.String(), lesson.TeachingMethod)
	}
	if lesson.SchedulingStatus != cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected SchedulingStatus %s but got %s", cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT, lesson.SchedulingStatus)
	}

	if !stringutil.SliceElementsMatch(lesson.TeacherIds, req.TeacherIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for teacher IDs, got %s", req.TeacherIds, lesson.TeacherIds)
	}

	learnerIds := make([]string, 0, len(req.StudentInfoList))
	for _, studentInfo := range req.StudentInfoList {
		learnerIds = append(learnerIds, studentInfo.StudentId)
	}
	actualLearnerIDs := make([]string, 0, len(lesson.LearnerMembers))
	for _, studentInfo := range lesson.LearnerMembers {
		actualLearnerIDs = append(actualLearnerIDs, studentInfo.StudentId)
	}
	if !stringutil.SliceElementsMatch(actualLearnerIDs, learnerIds) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected %s for learner IDs, got %s", learnerIds, actualLearnerIDs)
	}

	return StepStateToContext(ctx, stepState), nil
}
