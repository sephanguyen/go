package virtualclassroom

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) GetCurrentStateOfLiveLessonRoom(ctx context.Context, lessonID string) (*vpb.GetLiveLessonStateResponse, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetLiveLessonStateRequest{
		LessonId: lessonID,
	}

	res, err := vpb.NewVirtualClassroomReaderServiceClient(s.VirtualClassroomConn).
		GetLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *suite) GetCurrentStateOfLiveLessonRoomInBob(ctx context.Context, lessonID string) (*bpb.LiveLessonStateResponse, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.LiveLessonStateRequest{Id: lessonID}
	res, err := bpb.NewLessonReaderServiceClient(s.BobConn).
		GetLiveLessonState(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
