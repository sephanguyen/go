package virtualclassroom

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"google.golang.org/protobuf/proto"
)

func (s *suite) userDeleteALesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := lpb.DeleteLessonRequest{
		LessonId: stepState.CurrentLessonID,
	}
	ctx, err := s.createDeletedLessonSubscription(StepStateToContext(ctx, stepState))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createDeletedLessonSubscription: %w", err)
	}
	_, stepState.ResponseErr = lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).DeleteLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createDeletedLessonSubscription(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.FoundChanForJetStream = make(chan interface{}, 1)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.StartTime(time.Now()),
			nats.ManualAck(),
			nats.AckWait(2 * time.Second),
		},
	}
	handlerLessonDeletedSubscription := func(ctx context.Context, data []byte) (bool, error) {
		r := &bpb.EvtLesson{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return false, err
		}
		switch r.Message.(type) {
		case *bpb.EvtLesson_DeletedLessons_:
			stepState.DeletedLessonIDs = r.GetDeletedLessons().LessonIds
			stepState.FoundChanForJetStream <- r.Message
			return true, nil
		default:
			return false, fmt.Errorf("type message is invalid, expected EvtLesson_DeletedLessons_ but got: %T", r.Message)
		}
	}
	sub, err := s.JSM.Subscribe(constants.SubjectLessonDeleted, opts, handlerLessonDeletedSubscription)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	stepState.Subs = append(stepState.Subs, sub.JetStreamSub)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) recordedVideoWillBeDeletedInDBAndCloudStorage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(time.Second * 3)
	if stepState.CurrentLessonID == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stepState.CurrentLessonID must be not empty")
	}

	req := &vpb.GetRecordingByLessonIDRequest{
		LessonId: stepState.CurrentLessonID,
		Paging: &cpb.Paging{
			Limit: 2,
			Offset: &cpb.Paging_OffsetString{
				OffsetString: "",
			},
		},
	}

	res, err := vpb.NewLessonRecordingServiceClient(s.VirtualClassroomConn).GetRecordingByLessonID(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if res.TotalItems > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("recorded Video List of lesson %s must be non-existed", stepState.CurrentLessonID)
	}

	if ctx, err = s.checkExistedMedia(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return s.checkExistedResourceInCloud(ctx)
}

func (s *suite) checkExistedResourceInCloud(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, v := range stepState.RecordedVideos {
		path := filepath.Dir(v.Media.Resource) + "/"
		objects, err := s.CommonSuite.FileStore.GetObjectsWithPrefix(ctx, s.Cfg.Storage.Bucket, path, "")
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("FileStore.GetObjectsWithPrefix: %s %w", path, err)
		}
		if len(objects) > 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("all objects in path %s have to removed", path)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkExistedMedia(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ids, err := new(repo.MediaRepo).ListByIDs(ctx, s.LessonmgmtDBTrace, stepState.RecordedVideos.GetMediaIDs())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(ids) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("media list of lesson %s must be non-existed", stepState.CurrentLessonID)
	}
	return StepStateToContext(ctx, stepState), nil
}
