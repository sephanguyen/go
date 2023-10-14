package services

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandleEvent_LOCompleted(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	loStudyPlanItemRepo := &mock_repositories.MockLoStudyPlanItemRepo{}
	s := &StudentService{
		LoStudyPlanItemRepo: loStudyPlanItemRepo,
	}

	loID := idutil.ULIDNow()
	studyPlanItemID := idutil.ULIDNow()
	logs := generateStudentEventLog(studyPlanItemID, loID)
	testcases := []TestCase{
		{
			name:        "happy case",
			ctx:         context.Background(),
			expectedErr: nil,
			req: &epb.CreateStudentEventLogsRequest{
				StudentEventLogs: logs,
			},
			setup: func(ctx context.Context) {
				loStudyPlanItemRepo.On("UpdateCompleted", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "empty log",
			ctx:         context.Background(),
			expectedErr: nil,
			req:         &epb.CreateStudentEventLogsRequest{},
			setup: func(ctx context.Context) {
				loStudyPlanItemRepo.On("UpdateCompleted", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			err := s.HandleStudentEvent(ctx, testCase.req.(*epb.CreateStudentEventLogsRequest))
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}

func generateStudentEventLog(studyPlanItemID string, loID string) []*epb.StudentEventLog {
	SessionID := idutil.ULIDNow()
	StudyPlanItemID := studyPlanItemID
	LoID := loID

	logs := []*epb.StudentEventLog{
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-time.Hour)),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       SessionID,
				StudyPlanItemId: StudyPlanItemID,
				LoId:            LoID,
				Event:           "started",
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-45 * time.Minute)),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       SessionID,
				StudyPlanItemId: StudyPlanItemID,
				LoId:            LoID,
				Event:           "paused",
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now().Add(-15 * time.Minute)),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       SessionID,
				StudyPlanItemId: StudyPlanItemID,
				LoId:            LoID,
				Event:           "resumed",
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: timestamppb.New(time.Now()),
			Payload: &epb.StudentEventLogPayload{
				SessionId:       SessionID,
				StudyPlanItemId: StudyPlanItemID,
				LoId:            LoID,
				Event:           "completed",
			},
		},
	}

	return logs
}
