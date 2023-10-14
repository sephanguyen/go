package bob

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *suite) studentJoinLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = pb.NewClassClient(s.Conn).JoinLesson(contextWithToken(s, ctx), &pb.JoinLessonRequest{
		LessonId: stepState.CurrentLessonID,
	})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) StudentJoinLesson(ctx context.Context) (context.Context, error) {
	return s.studentJoinLesson(ctx)
}
func (s *suite) studentJoinLessonV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.Conn).JoinLesson(contextWithToken(s, ctx), &bpb.JoinLessonRequest{
		LessonId: stepState.CurrentLessonID,
	})

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) StudentJoinLessonV1(ctx context.Context) (context.Context, error) {
	return s.studentJoinLessonV1(ctx)
}
func (s *suite) StudentMustReceiveLessonRoomId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.JoinLessonResponse)
	if resp.RoomId == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing room id when student join live lesson")
	}
	stepState.RoomID = resp.RoomId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) MustReceiveLessonRoomIDAfterJoinLessonWhichSameCurrentRoomID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.JoinLessonResponse)
	if resp.RoomId == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing room id")
	}

	if stepState.RoomID != resp.RoomId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected room ID %s but got %s", stepState.RoomID, resp.RoomId)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) StudentMustReceiveLessonRoomID(ctx context.Context) (context.Context, error) {
	return s.StudentMustReceiveLessonRoomId(ctx)
}
func (s *suite) studentMustReceiveLessonRoomIDAndTokens(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*bpb.JoinLessonResponse)
	if resp.RoomId == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing room id when student join live lesson")
	}

	if resp.StmToken == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing rtm token")
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) StudentMustReceiveLessonRoomIdAndTokens(ctx context.Context) (context.Context, error) {
	return s.studentMustReceiveLessonRoomIDAndTokens(ctx)
}
