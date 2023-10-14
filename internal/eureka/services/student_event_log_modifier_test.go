package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/eureka/repositories"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestStudentEventLogModifier_CreateStudentEventLogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	studentEventLogRepo := new(mock_repositories.MockStudentEventLogRepo)
	mockDB := &mock_database.Ext{}
	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			req:         &epb.CreateStudentEventLogsRequest{},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "repo err case",
			ctx:         ctx,
			req:         &epb.CreateStudentEventLogsRequest{},
			expectedErr: status.Error(codes.Internal, errors.Wrap(pgx.ErrNoRows, "StudentEventLogRepo.Create").Error()),
			setup: func(ctx context.Context) {
				studentEventLogRepo.On("Create", ctx, mock.Anything, mock.Anything).Once().Return(pgx.ErrNoRows)
			},
		},
	}
	s := &StudentEventLogModifierService{
		StudentEventLogRepo: studentEventLogRepo,
		DB:                  mockDB,
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*epb.CreateStudentEventLogsRequest)
			_, err := s.CreateStudentEventLogs(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestStudentEventLogModifier_toStudentEventLogEntity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	userID := "userID1"
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("old cases", func(t *testing.T) {
		var output []*entities.StudentEventLog
		now := timestamppb.Now()
		input := []*epb.StudentEventLog{
			// Happy case
			{
				EventId:   "event-id-1",
				EventType: "test",
				Payload: &epb.StudentEventLogPayload{
					StudyPlanItemId: "study-plan-item-id-1",
				},
				CreatedAt: now,
			}}
		expected := []*entities.StudentEventLog{
			{
				ID:                 pgtype.Int4{Status: pgtype.Null},
				StudentID:          database.Text(userID),
				StudyPlanID:        pgtype.Text{Status: pgtype.Null},
				LearningMaterialID: pgtype.Text{Status: pgtype.Null},
				EventID:            database.Varchar("event-id-1"),
				EventType:          database.Varchar("test"),
				Payload: database.JSONB(&epb.StudentEventLogPayload{
					StudyPlanItemId: "study-plan-item-id-1",
				}),
				CreatedAt: database.Timestamptz(now.AsTime()),
			}}

		for _, tc := range input {
			rs, _ := toStudentEventLogEntity(ctx, tc)
			output = append(output, rs)
		}

		for idx, expect := range expected {
			assert.Equal(t, expect, output[idx])
		}
	})

	t.Run("add extra payload to payload JSONB", func(t *testing.T) {
		// arrange
		now := timestamppb.Now()
		input := &epb.StudentEventLog{
			EventId:   "event-id-1",
			EventType: "test",
			Payload: &epb.StudentEventLogPayload{
				StudyPlanItemId: "study-plan-item-id-1",
			},
			ExtraPayload: map[string]string{
				"props_a": "a",
				"props_b": "b",
			},
			CreatedAt: now,
		}
		expected := &entities.StudentEventLog{
			ID:                 pgtype.Int4{Status: pgtype.Null},
			StudentID:          database.Text(userID),
			StudyPlanID:        pgtype.Text{Status: pgtype.Null},
			LearningMaterialID: pgtype.Text{Status: pgtype.Null},
			EventID:            database.Varchar("event-id-1"),
			EventType:          database.Varchar("test"),
			Payload:            pgtype.JSONB{Status: pgtype.Present, Bytes: json.RawMessage(`{"props_a":"a","props_b":"b","study_plan_item_id":"study-plan-item-id-1"}`)},
			CreatedAt:          database.Timestamptz(now.AsTime()),
		}

		// act
		actual, err := toStudentEventLogEntity(ctx, input)

		// assert
		assert.Equal(t, expected, actual)
		assert.Nil(t, err)
	})
}
