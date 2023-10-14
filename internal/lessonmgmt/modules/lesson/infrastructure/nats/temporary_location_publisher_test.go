package nats

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLessonPublisher_PublishTemporaryLocationAssignment(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	jsm := new(mock_nats.JetStreamManagement)
	t.Run("happy case", func(t *testing.T) {
		lessonPublisher := LessonPublisher{}
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectEnrollmentStatusAssignmentCreated, mock.Anything).Once().Return("", nil)
		err := lessonPublisher.PublishTemporaryLocationAssignment(ctx, jsm, &npb.LessonReallocateStudentEnrollmentStatusEvent{
			StudentEnrollmentStatus: []*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo{
				{
					StudentId:        "s1",
					LocationId:       "l1",
					StartDate:        timestamppb.Now(),
					EndDate:          timestamppb.Now(),
					EnrollmentStatus: npb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY,
				},
			},
		})
		assert.NoError(t, err)
	})
	t.Run("failed", func(t *testing.T) {
		lessonPublisher := LessonPublisher{}
		jsm.On("PublishAsyncContext", mock.Anything, constants.SubjectEnrollmentStatusAssignmentCreated, mock.Anything).Once().Return("", errors.New("something went wrong"))
		err := lessonPublisher.PublishTemporaryLocationAssignment(ctx, jsm, &npb.LessonReallocateStudentEnrollmentStatusEvent{})
		assert.Error(t, err)
	})
}
