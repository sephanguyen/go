package bob

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	pbb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) teacherEndOneOfTheLiveLessonV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if reflect.TypeOf(stepState.Response) != reflect.TypeOf(&pb.RetrieveLiveLessonResponse{}) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response need to be RetrieveLiveLessonResponse")
	}
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	if len(rsp.Lessons) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response does not have lessons")
	}
	token, err := s.generateExchangeToken(stepState.CurrentTeacherID, entities.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err

	}

	lessonID := ""
	for _, lesson := range rsp.Lessons {
		if lesson.Teacher[0].UserId == stepState.CurrentTeacherID {
			lessonID = lesson.LessonId
		}
	}
	if lessonID == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot find lesson of current teacher")
	}
	req := &pbb.EndLiveLessonRequest{
		LessonId: lessonID,
	}
	stepState.CurrentLessonID = lessonID
	stepState.Request = req
	ctx, err = s.createEndLiveLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEndLiveLessonSubscription: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pbb.NewClassModifierServiceClient(s.Conn).EndLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theEndedLessonMustHaveStatusCompletedV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	found := false
	for _, lesson := range rsp.Lessons {
		if lesson.LessonId == stepState.CurrentLessonID && lesson.Status == pb.LESSON_STATUS_COMPLETED {
			found = true
			break
		}
	}
	if !found {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s has status %s", stepState.CurrentLessonID, pb.LESSON_STATUS_COMPLETED.String())
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustUpdateLessonEndAtTimeV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pbb.EndLiveLessonRequest)
	query := "SELECT end_at FROM lessons l WHERE l.lesson_id = $1"
	var endDate pgtype.Timestamptz
	if err := s.DBPostgres.QueryRow(ctx, query, req.LessonId).Scan(&endDate); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if endDate.Time.Unix() > time.Now().Unix() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson is not ended")
	}
	return StepStateToContext(ctx, stepState), nil
}
