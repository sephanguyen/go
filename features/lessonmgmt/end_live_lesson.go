package lessonmgmt

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/manabie-com/backend/features/helper"
	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
)

func (s *Suite) studentRetrieveLiveLessonWithStartTimeAndEndTime(ctx context.Context, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = stepState.CurrentStudentID
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	token, err := s.CommonSuite.GenerateExchangeToken(stepState.CurrentStudentID, entities_bob.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From: &types.Timestamp{Seconds: startDate.Unix()},
		To:   &types.Timestamp{Seconds: endDate.Unix()},
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Connections.BobConn).RetrieveLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) teacherRetrieveLiveLessonWithStartTimeAndEndTime(ctx context.Context, startTimeString, endTimeString string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startDate, err := time.Parse(time.RFC3339, startTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	endDate, err := time.Parse(time.RFC3339, endTimeString)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	token, err := s.CommonSuite.GenerateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := &pb.RetrieveLiveLessonRequest{
		Pagination: &pb.Pagination{
			Limit: 100,
			Page:  1,
		},
		From: &types.Timestamp{Seconds: startDate.Unix()},
		To:   &types.Timestamp{Seconds: endDate.Unix()},
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseClient(s.Connections.BobConn).RetrieveLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) teacherEndOneOfTheLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if reflect.TypeOf(stepState.Response) != reflect.TypeOf(&pb.RetrieveLiveLessonResponse{}) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response need to be RetrieveLiveLessonResponse")
	}
	rsp := stepState.Response.(*pb.RetrieveLiveLessonResponse)
	if len(rsp.Lessons) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("response does not have lessons")
	}
	token, err := s.CommonSuite.GenerateExchangeToken(stepState.CurrentTeacherID, entities_bob.UserGroupTeacher)
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
	req := &pb.EndLiveLessonRequest{
		LessonId: lessonID,
	}
	stepState.CurrentLessonID = lessonID
	stepState.Request = req
	ctx, err = s.createEndLiveLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createEndLiveLessonSubscription: %w", err)
	}
	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.BobConn).EndLiveLesson(helper.GRPCContext(ctx, "token", token), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createEndLiveLessonSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonEndedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &pb.EvtLesson{}
		err := r.Unmarshal(data)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *pb.EvtLesson_EndLiveLesson_:
			if r.GetEndLiveLesson().LessonId == stepState.CurrentLessonID && r.GetEndLiveLesson().UserId == stepState.CurrentTeacherID {
				stepState.FoundChanForJetStream <- r.Message
				return false, nil
			}
		}
		return false, errors.New("StudentID not equal leanerID")
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonUpdated, opts, handlerLessonEndedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) bobMustUpdateLessonEndAtTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.EndLiveLessonRequest)
	query := "SELECT end_at FROM lessons l WHERE l.lesson_id = $1"
	var endDate pgtype.Timestamptz
	if err := s.BobDB.QueryRow(ctx, query, req.LessonId).Scan(&endDate); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if endDate.Time.Unix() > time.Now().Unix() {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson is not ended")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theEndedLessonMustHaveStatusCompleted(ctx context.Context) (context.Context, error) {
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
