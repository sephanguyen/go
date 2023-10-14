package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"
	commonpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	CancelState  = "cancel"
	PublishState = "publish"
)

func (s *suite) someExistingStatusLessons(ctx context.Context, status string) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)
	stepState.CurrentTeachingMethod = "individual"
	timezone := LoadLocalLocation()
	lessonIDsForBulkAction := make([]string, 0, 5)
	var lessonStatus lpb.LessonStatus

	switch status {
	case "published":
		lessonStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_PUBLISHED
	case "completed":
		lessonStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_COMPLETED
	case "canceled":
		lessonStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_CANCELED
	case "draft":
		lessonStatus = lpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT
	}

	// create 10 lessons
	for i := 0; i < 10; i++ {
		req := s.CommonSuite.UserCreateALessonRequestWithMissingFieldsInLessonmgmt(StepStateToContext(ctx, stepState), commonpb.LessonTeachingMedium_LESSON_TEACHING_MEDIUM_OFFLINE)
		req.StartTime = timestamppb.New(time.Date(2022, 12, i+1, 22, 0, 0, 0, timezone))
		req.EndTime = timestamppb.New(time.Date(2022, 12, i+1, 23, 45, 0, 0, timezone))
		req.SchedulingStatus = lessonStatus

		// for lesson draft scenario, create 2 draft lessons not ready for publish status
		if i > 7 && lessonStatus == lpb.LessonStatus_LESSON_SCHEDULING_STATUS_DRAFT {
			req.TeacherIds = []string{}
		}

		ctx, _ = s.CommonSuite.UserCreateALessonWithRequestInLessonmgmt(ctx, req)
		stepState = StepStateFromContext(ctx)
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error encountered when creating lessons: %w", stepState.ResponseErr)
		}

		lessonIDsForBulkAction = append(lessonIDsForBulkAction, stepState.CurrentLessonID)
	}
	stepState.LessonIDs = lessonIDsForBulkAction

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetLessonIDsforBulkAction(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	timezone := LoadLocalLocation()

	req := &cpb.GetLessonIDsForBulkStatusUpdateRequest{
		LocationId: stepState.CenterIDs[len(stepState.CenterIDs)-1],
		StartDate:  timestamppb.New(time.Date(2022, 12, 1, 0, 0, 0, 0, timezone)),
		EndDate:    timestamppb.New(time.Date(2022, 12, 13, 23, 59, 59, 0, timezone)),
		StartTime:  timestamppb.New(time.Date(2022, 12, 1, 22, 0, 0, 0, timezone)),
		EndTime:    timestamppb.New(time.Date(2022, 12, 1, 23, 45, 0, 0, timezone)),
		Timezone:   timezone.String(),
	}

	switch action {
	case CancelState:
		req.Action = lpb.LessonBulkAction_LESSON_BULK_ACTION_CANCEL
	case PublishState:
		req.Action = lpb.LessonBulkAction_LESSON_BULK_ACTION_PUBLISH
	}
	stepState.Request = req

	ctx = s.signedCtx(StepStateToContext(ctx, stepState))
	stepState.Response, stepState.ResponseErr = cpb.NewLessonReaderServiceClient(s.CalendarConn).
		GetLessonIDsForBulkStatusUpdate(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnedLessonIDsForBulkActionAreExpected(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := stepState.Request.(*cpb.GetLessonIDsForBulkStatusUpdateRequest)
	response := stepState.Response.(*cpb.GetLessonIDsForBulkStatusUpdateResponse)
	requestStartDate := request.StartDate.AsTime()
	requestEndDate := request.EndDate.AsTime()

	for _, lessonIDDetail := range response.GetLessonIdsDetails() {
		numberOfLessonIDs := uint32(len(lessonIDDetail.GetLessonIds()))

		// checking lesson status on header level
		if action == CancelState {
			if lessonIDDetail.SchedulingStatus != commonpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED &&
				lessonIDDetail.SchedulingStatus != commonpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson status in header is not completed or published for bulk cancel: %v", lessonIDDetail.SchedulingStatus)
			}

			if lessonIDDetail.LessonsCount != numberOfLessonIDs {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the number of total lessons %v is not equal to the lesson IDs %v", lessonIDDetail.LessonsCount, lessonIDDetail)
			}

			if lessonIDDetail.ModifiableLessonsCount != numberOfLessonIDs {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the number of total modifiable lessons %v is not equal to the lesson IDs %v", lessonIDDetail.LessonsCount, lessonIDDetail)
			}
		} else if action == PublishState {
			if lessonIDDetail.SchedulingStatus != commonpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson status in header is not draft for bulk publish: %v", lessonIDDetail.SchedulingStatus)
			}

			if lessonIDDetail.ModifiableLessonsCount != numberOfLessonIDs {
				return StepStateToContext(ctx, stepState), fmt.Errorf("the number of total modifiable lessons %v is not equal to the lesson IDs %v", lessonIDDetail.LessonsCount, lessonIDDetail)
			}
		}

		// verify each lesson ID
		for _, lessonID := range lessonIDDetail.GetLessonIds() {
			ctx = s.getLessonByID(ctx, lessonID)
			if stepState.ResponseErr != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get lesson id %s: %w", lessonID, stepState.ResponseErr)
			}

			stepState = StepStateFromContext(ctx)
			lessonDetail := stepState.Response.(*lpb.RetrieveLessonByIDResponse)
			lessonStartTime := lessonDetail.Lesson.StartTime.AsTime()
			lessonEndTime := lessonDetail.Lesson.EndTime.AsTime()

			if requestStartDate.After(lessonStartTime) && !requestStartDate.Equal(lessonStartTime) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s start date %v is found after the request start date %v", lessonID, lessonStartTime, requestStartDate)
			}

			if requestEndDate.Before(lessonEndTime) && !requestEndDate.Equal(lessonEndTime) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s end date %v is found before the request end date %v", lessonID, lessonEndTime, requestEndDate)
			}

			// checking lesson status on detail level
			if action == CancelState {
				if lessonDetail.Lesson.SchedulingStatus != commonpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED &&
					lessonDetail.Lesson.SchedulingStatus != commonpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED {
					return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %v status is not completed or published: %s", lessonID, lessonDetail.Lesson.SchedulingStatus)
				}
			} else if action == PublishState {
				if lessonDetail.Lesson.SchedulingStatus != commonpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_DRAFT {
					return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %v status is not draft: %s", lessonID, lessonDetail.Lesson.SchedulingStatus)
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getLessonByID(ctx context.Context, lessonID string) context.Context {
	stepState := StepStateFromContext(ctx)

	req := &lpb.RetrieveLessonByIDRequest{
		LessonId: lessonID,
	}

	ctx = s.signedCtx(StepStateToContext(ctx, stepState))
	stepState.Response, stepState.ResponseErr = lpb.NewLessonReaderServiceClient(s.LessonMgmtConn).
		RetrieveLessonByID(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState)
}
