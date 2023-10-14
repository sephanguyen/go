package learnosity

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/learnosity"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	mock_learnosity "github.com/manabie-com/backend/mock/golibs/learnosity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSessionRepo_GetSessionStatuses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	userID := "user_id"
	status := []string{string(learnosity.SessionStatusCompleted)}
	endpoint := learnosity.EndpointDataAPISessionsStatuses
	completedAt := time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC)
	security := learnosity.Security{
		ConsumerKey:    "consumer_key",
		Domain:         "domain",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         interceptors.UserIDFromContext(ctx),
		ConsumerSecret: "secret",
	}

	t.Run("get all statuses successfully", func(t *testing.T) {
		// Arrange
		mockHTTP := &mock_learnosity.HTTP{}
		mockDataAPI := &mock_learnosity.DataAPI{}
		repo := NewSessionRepo(mockHTTP, mockDataAPI)
		asmIDs := []string{"LM_1", "LM_2", "LM3", "LM4"}
		request := learnosity.Request{
			"activity_id": asmIDs,
			"user_id":     []string{userID},
			"status":      status,
		}
		expectedSessions := sliceutils.Map(asmIDs, func(a string) domain.Session {
			return domain.Session{
				ID:           idutil.ULIDNow(),
				AssessmentID: a,
				UserID:       userID,
				Status:       domain.SessionStatusCompleted,
				CompletedAt:  &completedAt,
			}
		})
		learnosityDataArr := sliceutils.Map(expectedSessions, func(s domain.Session) map[string]any {
			return map[string]any{
				"session_id":   s.ID,
				"user_id":      s.UserID,
				"activity_id":  s.AssessmentID,
				"status":       "Completed",
				"dt_completed": completedAt,
			}
		})
		dataRaw, _ := json.Marshal(learnosityDataArr)
		mockDataAPI.On("RequestIterator", mock.Anything, mockHTTP, endpoint, security, request).
			Once().
			Return([]learnosity.Result{
				{
					Meta: map[string]any{"records": float64(len(learnosityDataArr))},
					Data: dataRaw,
				},
			}, nil)

		// Act
		statuses, err := repo.GetSessionStatuses(ctx, security, request)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedSessions, statuses)
		mock.AssertExpectationsForObjects(t, mockDataAPI)
	})

	t.Run("get all statuses failed", func(t *testing.T) {
		// Arrange
		mockHTTP := &mock_learnosity.HTTP{}
		mockDataAPI := &mock_learnosity.DataAPI{}
		repo := NewSessionRepo(mockHTTP, mockDataAPI)
		lmIDs := []string{"LM_1"}
		request := learnosity.Request{
			"activity_id": lmIDs,
			"user_id":     []string{userID},
			"status":      status,
		}
		apiErr := fmt.Errorf("%s", "some err")
		expectedErr := errors.NewLearnosityError("LearnositySessionRepo.GetSessionStatuses", apiErr)
		mockDataAPI.On("RequestIterator", mock.Anything, mockHTTP, endpoint, security, request).
			Once().
			Return(nil, apiErr)

		// Act
		statuses, err := repo.GetSessionStatuses(ctx, security, request)

		// Assert
		assert.Nil(t, statuses)
		assert.Equal(t, expectedErr, err)
		mock.AssertExpectationsForObjects(t, mockDataAPI)
	})
}

func TestSessionRepo_GetSessionResponses(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockHTTP := &mock_learnosity.HTTP{}
	mockDataAPI := &mock_learnosity.DataAPI{}
	repo := NewSessionRepo(mockHTTP, mockDataAPI)

	security := learnosity.Security{
		ConsumerKey:    "consumer_key",
		Domain:         "domain",
		Timestamp:      learnosity.FormatUTCTime(time.Now()),
		UserID:         interceptors.UserIDFromContext(ctx),
		ConsumerSecret: "secret",
	}
	completedAt := time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		Name             string
		Ctx              context.Context
		Request          any
		Setup            func(ctx context.Context)
		ExpectedResponse any
		ExpectedError    error
	}{
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user_id"),
			Request: learnosity.Request{
				"activity_id": []string{"activity_id"},
				"user_id":     []string{"user_id"},
			},
			Setup: func(ctx context.Context) {
				dataArr := []any{
					map[string]any{
						"session_id":   "session_id",
						"max_score":    8,
						"score":        4,
						"status":       "Completed",
						"dt_started":   time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC),
						"dt_completed": completedAt,
					},
				}
				dataRaw, _ := json.Marshal(dataArr)
				mockDataAPI.On("RequestIterator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return([]learnosity.Result{
					{
						Meta: map[string]any{"records": float64(1)},
						Data: dataRaw,
					},
				}, nil)
			},
			ExpectedResponse: domain.Sessions{
				{
					ID:          "session_id",
					MaxScore:    8,
					GradedScore: 4,
					Status:      domain.SessionStatusCompleted,
					CreatedAt:   time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC),
					CompletedAt: &completedAt,
				},
			},
			ExpectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.Setup(tc.Ctx)
			resp, err := repo.GetSessionResponses(tc.Ctx, security, tc.Request.(learnosity.Request))
			if tc.ExpectedError != nil {
				assert.Equal(t, tc.ExpectedError.Error(), err.Error())
			} else {
				assert.Equal(t, resp, tc.ExpectedResponse.(domain.Sessions))
			}
		})
	}
}
