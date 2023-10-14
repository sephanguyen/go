package producers

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/protobuf/proto"
)

type LessonProducer struct {
	JSM nats.JetStreamManagement
}

// TODO: move to infrastructure
//
//nolint:interfacer
func (l *LessonProducer) PublishLessonEvt(ctx context.Context, msg *bpb.EvtLesson) error {
	var subject string
	switch msg.Message.(type) {
	case *bpb.EvtLesson_CreateLessons_:
		subject = constants.SubjectLessonCreated
	case *bpb.EvtLesson_DeletedLessons_:
		subject = constants.SubjectLessonDeleted
	default:
		subject = constants.SubjectLessonUpdated
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	msgID, err := l.JSM.PublishAsyncContext(ctx, subject, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("PublishLessonEvt JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}
