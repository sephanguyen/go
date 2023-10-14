package lessonmgmt

import (
	"context"
	"fmt"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/multierr"
)

func (s *Suite) returnsValidInformationForStudentBroadcast(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.JoinLessonResponse)

	var err error

	if rsp.RoomId == "" {
		err = multierr.Append(err, fmt.Errorf("did not return valid room id"))
	}

	if rsp.StreamToken == "" {
		err = multierr.Append(err, fmt.Errorf("did not return stream token"))
	}

	if rsp.WhiteboardToken == "" {
		err = multierr.Append(err, fmt.Errorf("did not return whiteboard token"))
	}

	if rsp.AgoraAppId == "" {
		err = multierr.Append(err, fmt.Errorf("did not return valid Agora AppId"))
	}

	if rsp.WhiteboardAppId == "" {
		err = multierr.Append(err, fmt.Errorf("did not return valid Whiteboard AppId"))
	}

	return StepStateToContext(ctx, stepState), err
}
