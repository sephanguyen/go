package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

type LessonPublisher struct {
}

func (l *LessonPublisher) PublishTemporaryLocationAssignment(ctx context.Context, jsm nats.JetStreamManagement, msg *npb.LessonReallocateStudentEnrollmentStatusEvent) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	msgID, err := jsm.PublishAsyncContext(ctx, constants.SubjectEnrollmentStatusAssignmentCreated, data)
	if err != nil {
		return nats.HandlePushMsgFail(ctx, fmt.Errorf("JSM.PublishAsyncContext failed, msgID: %s, %w", msgID, err))
	}
	return nil
}
