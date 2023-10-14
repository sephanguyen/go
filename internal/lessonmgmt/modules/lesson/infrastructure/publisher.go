package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

type LessonPublisher interface {
	PublishTemporaryLocationAssignment(ctx context.Context, jsm nats.JetStreamManagement, msg *npb.LessonReallocateStudentEnrollmentStatusEvent) error
}
