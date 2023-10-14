package domain_test

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCurrentMaterial_IsValid(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name            string
		currentMaterial *domain.CurrentMaterial
		hasError        bool
	}{
		{
			name: "full fields with video state",
			currentMaterial: &domain.CurrentMaterial{
				MediaID:   "media-1",
				UpdatedAt: time.Now(),
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(time.Duration(20)),
					PlayerState: domain.PlayerStateEnded,
				},
			},
			hasError: false,
		},
		{
			name: "full fields with audio state",
			currentMaterial: &domain.CurrentMaterial{
				MediaID:   "media-1",
				UpdatedAt: time.Now(),
				AudioState: &domain.AudioState{
					CurrentTime: domain.Duration(time.Duration(20)),
					PlayerState: domain.PlayerStateEnded,
				},
			},
			hasError: false,
		},
		{
			name: "missing mediaID",
			currentMaterial: &domain.CurrentMaterial{
				MediaID:   "",
				UpdatedAt: time.Now(),
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(time.Duration(20)),
					PlayerState: domain.PlayerStateEnded,
				},
			},
			hasError: true,
		},
		{
			name: "updated_at is zero",
			currentMaterial: &domain.CurrentMaterial{
				MediaID:   "media-1",
				UpdatedAt: time.Time{},
				VideoState: &domain.VideoState{
					CurrentTime: domain.Duration(time.Duration(20)),
					PlayerState: domain.PlayerStateEnded,
				},
			},
			hasError: true,
		},
	}

	for i, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.currentMaterial.IsValid()
			if tcs[i].hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(
				t,
			)
		})
	}
}

func TestVirtualClassroomState_NewLearnerState(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                 string
		states               *domain.LessonMemberStates
		isValid              bool
		expectedLearnerState *domain.LearnerState
	}{
		{
			name: "full fields",
			states: &domain.LessonMemberStates{
				&domain.LessonMemberState{
					LessonID:         "lesson-id-1",
					UserID:           "user-id-1",
					AttendanceStatus: "test",
					AttendanceRemark: "test",
					CourseID:         "course-id-1",
				},
			},
			expectedLearnerState: &domain.LearnerState{
				UserID: "user-id-1",
				//TODO: Update more data once these states are implemented
				HandsUp:       &domain.UserHandsUp{},
				Annotation:    &domain.UserAnnotation{},
				PollingAnswer: &domain.UserPollingAnswer{},
				Chat:          &domain.UserChat{},
			},

			isValid: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			learnerState := domain.NewLearnerState("user-id-1", *tc.states)
			assert.EqualValues(t, tc.expectedLearnerState, learnerState)

			mock.AssertExpectationsForObjects(
				t,
			)
		})
	}
}

func TestVirtualClassroomState_NewUserState(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name                 string
		states               *domain.LessonMemberStates
		isValid              bool
		expectedLearnerState *domain.UserStates
	}{
		{
			name: "full fields",
			states: &domain.LessonMemberStates{
				&domain.LessonMemberState{
					LessonID:         "lesson-id-1",
					UserID:           "user-id-1",
					AttendanceStatus: "test",
					AttendanceRemark: "test",
					CourseID:         "course-id-1",
				},
			},
			expectedLearnerState: &domain.UserStates{
				LearnersState: []*domain.LearnerState{
					{
						UserID:        "user-id-1",
						HandsUp:       &domain.UserHandsUp{},
						Annotation:    &domain.UserAnnotation{},
						PollingAnswer: &domain.UserPollingAnswer{},
						Chat:          &domain.UserChat{},
					},
				},
			},

			isValid: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			learnerState := domain.NewUserState(*tc.states)
			assert.EqualValues(t, tc.expectedLearnerState, learnerState)

			mock.AssertExpectationsForObjects(
				t,
			)
		})
	}
}
