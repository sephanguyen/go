package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/features/helper"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) userDeleteLessonRecurring(ctx context.Context, lessonIndex string, method string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonIndexNum, err := strconv.Atoi(lessonIndex)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := lpb.DeleteLessonRequest{
		LessonId: stepState.LessonIDs[lessonIndexNum],
	}
	switch method {
	case "one_time":
		req.SavingOption = &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_ONE_TIME}
	case "recurring":
		req.SavingOption = &lpb.DeleteLessonRequest_SavingOption{Method: lpb.CreateLessonSavingMethod_CREATE_LESSON_SAVING_METHOD_RECURRENCE}
	}
	ctx, err = s.createDeletedLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createDeletedLessonSubscription: %w", err)
	}
	_, stepState.ResponseErr = lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).DeleteLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &req)

	return StepStateToContext(ctx, stepState), nil
}

func getLessonByLessonListStr(ctx context.Context, lessonListStr string) []string {
	stepState := StepStateFromContext(ctx)
	lessons := make([]string, 0)

	for i := 0; i < len(stepState.LessonIDs); i++ {
		for _, y := range strings.Split(lessonListStr, ",") {
			if fmt.Sprint(i) == y {
				lessons = append(lessons, stepState.LessonIDs[i])
			}
		}
	}
	return lessons
}

func (s *Suite) UserSubmitANewLessonReportFromLessonRecurring(ctx context.Context, lessonListStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessons := getLessonByLessonListStr(ctx, lessonListStr)

	for _, v := range lessons {
		stepState.CurrentLessonID = v
		if ctx, err := s.UserSubmitANewLessonReport(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userNoLongerSeesTheLessonRecurring(ctx context.Context, lessonListStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessons := getLessonByLessonListStr(ctx, lessonListStr)

	for _, v := range lessons {
		stepState.CurrentLessonID = v
		if ctx, err := s.userNoLongerSeesTheLesson(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userNoLongerSeesTheLessonReportRecurring(ctx context.Context, lessonListStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessons := getLessonByLessonListStr(ctx, lessonListStr)

	for _, v := range lessons {
		stepState.CurrentLessonID = v
		if ctx, err := s.userNoLongerSeesTheLessonReport(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userStillSeesTheLessonReportRecurring(ctx context.Context, lessonListStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessons := getLessonByLessonListStr(ctx, lessonListStr)

	for _, v := range lessons {
		stepState.CurrentLessonID = v
		if ctx, err := s.userStillSeesTheLessonReport(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userStillSeesTheLessonRecurring(ctx context.Context, lessonListStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessons := getLessonByLessonListStr(ctx, lessonListStr)

	for _, v := range lessons {
		stepState.CurrentLessonID = v
		if ctx, err := s.userStillSeesTheLesson(ctx); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("%w", err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lockLessons(ctx context.Context, lockedLessons string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessons := getLessonByLessonListStr(ctx, lockedLessons)

	for _, v := range lessons {
		stepState.CurrentLessonID = v
		ctx, err := s.lockLesson(ctx, "true")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
