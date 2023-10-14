package bob

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
)

func (s *suite) teacherJoinLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).JoinLesson(contextWithToken(s, ctx), &pb.JoinLessonRequest{
		LessonId: stepState.CurrentLessonID,
	})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) TeacherJoinLesson(ctx context.Context) (context.Context, error) {
	return s.teacherJoinLesson(ctx)
}
func (s *suite) returnStreamToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.JoinLessonResponse)
	if rsp.StreamToken == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob did not return stream token")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnWhiteboardToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.JoinLessonResponse)
	if rsp.WhiteboardToken == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob did not return whiteboard token")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnValidRoomID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.JoinLessonResponse)
	if rsp.RoomId == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob did not return valid room id")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsValidInformationForBroadcast(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.returnStreamToken(ctx)
	ctx, err2 := s.returnWhiteboardToken(ctx)
	ctx, err3 := s.returnValidRoomID(ctx)

	return ctx, multierr.Combine(err1, err2, err3)
}
func (s *suite) ReturnsValidInformationForBroadcast(ctx context.Context) (context.Context, error) {
	return s.returnsValidInformationForBroadcast(ctx)
}
func (s *suite) returnsEmptyWhiteBoardToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.JoinLessonResponse)
	if rsp.WhiteboardToken != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return empty whiteboard token")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsEmptyRoomId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*pb.JoinLessonResponse)
	if rsp.WhiteboardToken != "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("bob must return empty room id")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsValidInformationForSubscribe(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.returnStreamToken(ctx)
	ctx, err2 := s.returnsEmptyWhiteBoardToken(ctx)
	ctx, err3 := s.returnsEmptyRoomId(ctx)
	return ctx, multierr.Combine(err1, err2, err3)

}
func (s *suite) teacherJoinLessonV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.Conn).JoinLesson(contextWithToken(s, ctx), &bpb.JoinLessonRequest{
		LessonId: stepState.CurrentLessonID,
	})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsValidInformationForBroadcastV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.JoinLessonResponse)

	var err error

	if rsp.RoomId == "" {
		err = multierr.Append(err, fmt.Errorf("bob did not return valid room id"))
	}

	if rsp.StreamToken == "" {
		err = multierr.Append(err, fmt.Errorf("bob did not return stream token"))
	}

	if rsp.WhiteboardToken == "" {
		err = multierr.Append(err, fmt.Errorf("bob did not return whiteboard token"))
	}

	if rsp.VideoToken == "" {
		err = multierr.Append(err, fmt.Errorf("bob did not return valid video token"))
	}

	if rsp.AgoraAppId == "" {
		err = multierr.Append(err, fmt.Errorf("bob did not return valid Agora AppId"))
	}

	if rsp.WhiteboardAppId == "" {
		err = multierr.Append(err, fmt.Errorf("bob did not return valid Whiteboard AppId"))
	}

	if rsp.ScreenRecordingToken == "" {
		err = multierr.Append(err, fmt.Errorf("bob did not return valid Share Screen For Recording Token"))
	}

	return StepStateToContext(ctx, stepState), err
}
