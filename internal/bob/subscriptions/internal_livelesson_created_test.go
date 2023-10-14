package subscriptions

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	mock_repositories "github.com/manabie-com/backend/mock/bob/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	bpb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func newLiveLessonCreatedSub() *InternalLiveLessonCreated {
	lessonRepo := new(mock_repositories.MockLessonRepo)
	whiteboardSvc := new(mock_whiteboard.MockService)
	mockDB := &mock_database.Ext{}

	return &InternalLiveLessonCreated{
		Logger:        zap.NewNop(),
		DB:            mockDB,
		LessonRepo:    lessonRepo,
		WhiteboardSvc: whiteboardSvc,
	}
}

type testCaseLessons struct {
	name            string
	event           interface{}
	setup           func(t *testing.T, s *InternalLiveLessonCreated, event interface{})
	expectedAckAble bool
	hasError        bool
}

func TestHandleCreateRoom(t *testing.T) {
	t.Parallel()

	testCases := []testCaseLessons{
		{
			name: "live lesson create room successfully",
			event: &bpb.EvtLesson_CreateLessons{
				Lessons: []*bpb.EvtLesson_Lesson{
					{
						Name:       "lesson-name",
						LessonId:   "lesson-1",
						LearnerIds: []string{"learner1", "learner2"},
					},
				},
			},
			setup: func(t *testing.T, s *InternalLiveLessonCreated, event interface{}) {
				s.WhiteboardSvc.(*mock_whiteboard.MockService).On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     "lesson-1",
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{
					UUID: "room-id-1",
				}, nil)
				s.LessonRepo.(*mock_repositories.MockLessonRepo).On("UpdateRoomID", mock.Anything, mock.Anything, database.Text("lesson-1"), database.Text("room-id-1")).Once().Return(nil)
			},
			expectedAckAble: true,
			hasError:        false,
		},
		{
			name: "live lesson create room fail when missing lesson_id",
			event: &bpb.EvtLesson_CreateLessons{
				Lessons: []*bpb.EvtLesson_Lesson{
					{
						Name:       "lesson-name",
						LearnerIds: []string{"learner1", "learner2"},
					},
				},
			},
			setup: func(t *testing.T, s *InternalLiveLessonCreated, event interface{}) {
				s.WhiteboardSvc.(*mock_whiteboard.MockService).On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{}, errors.New("nothing"))
				s.LessonRepo.(*mock_repositories.MockLessonRepo).On("UpdateRoomID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
			},
			expectedAckAble: false,
			hasError:        true,
		},
		{
			name: "live lesson create room fail when update room fail",
			event: &bpb.EvtLesson_CreateLessons{
				Lessons: []*bpb.EvtLesson_Lesson{
					{
						Name:       "lesson-name",
						LessonId:   "lesson-2",
						LearnerIds: []string{"learner1", "learner2"},
					},
				},
			},
			setup: func(t *testing.T, s *InternalLiveLessonCreated, event interface{}) {
				s.WhiteboardSvc.(*mock_whiteboard.MockService).On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     "lesson-2",
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{
					UUID: "room-id-1",
				}, nil)
				s.LessonRepo.(*mock_repositories.MockLessonRepo).On("UpdateRoomID", mock.Anything, mock.Anything, database.Text("lesson-2"), database.Text("room-id-1")).Once().Return(errors.New("cannot update lesson"))
			},
			expectedAckAble: false,
			hasError:        true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := newLiveLessonCreatedSub()
			tc.setup(t, s, tc.event)
			ackAble, err := s.handle(context.Background(), tc.event.(*bpb.EvtLesson_CreateLessons))
			if ackAble != tc.expectedAckAble {
				t.Errorf("unexpected ackAble: got: %v, want: %v", ackAble, tc.expectedAckAble)
			}
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
