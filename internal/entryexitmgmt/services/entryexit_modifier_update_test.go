package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	eepb "github.com/manabie-com/backend/pkg/manabuf/entryexitmgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestEntryExitModifierService_UpdateEntryExit(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// testTime will be used by ordinary tests
	// currentTime should be used by the tests that compares the time with current time
	currentTime := time.Now().UTC()
	testTime := time.Now().UTC()
	if testTime.Hour() == 0 {
		testTime = testTime.AddDate(0, 0, -1)
		testTime = testTime.Add(5 * time.Hour)
	}

	s, testSetup := initTestSetupAndServiceForCreateUpdate()

	entryRequestWithNotify := func(notify bool) *eepb.UpdateEntryExitRequest {
		return &eepb.UpdateEntryExitRequest{
			EntryExitPayload: &eepb.EntryExitPayload{
				StudentId:     "student-exist",
				EntryDateTime: timestamppb.New(testTime),
				ExitDateTime:  nil,
				NotifyParents: notify,
			},
			EntryexitId: 1,
		}
	}
	entryExitRequestWithNotify := func(notify bool) *eepb.UpdateEntryExitRequest {
		return &eepb.UpdateEntryExitRequest{
			EntryExitPayload: &eepb.EntryExitPayload{
				StudentId:     "student-exist",
				EntryDateTime: timestamppb.New(testTime.Add(-1 * time.Hour)),
				ExitDateTime:  timestamppb.New(testTime),
				NotifyParents: notify,
			},
			EntryexitId: 1,
		}
	}
	successResponseNotified := func(notify bool) *eepb.UpdateEntryExitResponse {
		return &eepb.UpdateEntryExitResponse{
			Successful:     true,
			ParentNotified: notify,
		}
	}

	testCases := []TestCase{
		{
			name:         "Happy case Update Entry Record",
			ctx:          ctx,
			req:          entryRequestWithNotify(false),
			expectedResp: successResponseNotified(false),
			setup:        testSetup.UpdateSetup(),
		},
		{
			name:         "Happy case Update both Entry and Exit Record",
			ctx:          ctx,
			req:          entryExitRequestWithNotify(false),
			expectedResp: successResponseNotified(false),
			setup:        testSetup.UpdateSetup(),
		},
		{
			name:         "Happy case Update Only Entry Date Record with notification",
			ctx:          ctx,
			req:          entryRequestWithNotify(true),
			expectedResp: successResponseNotified(true),
			setup:        testSetup.UpdateWithNotifSetup(),
		},
		{
			name:         "Happy case Update Entry and Exit Record with notification",
			ctx:          ctx,
			req:          entryExitRequestWithNotify(true),
			expectedResp: successResponseNotified(true),
			setup:        testSetup.UpdateWithNotifSetup(),
		},
		// failed to retrieve student
		{
			name: "cannot retrieve if student id not existing in mockDB",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "not-exist-id",
					EntryDateTime: timestamppb.New(testTime.Add(-1 * time.Hour)),
					ExitDateTime:  timestamppb.New(testTime),
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "student id does not exist"),
			setup:       testSetup.StudentNotExistSetup(),
		},
		{
			name: "student id is empty",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId: "",
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "student id cannot be empty"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "student entry exit id is not valid",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId: "exist",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student entry exit id must be valid"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "student entry exit id is zero",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId: "exist",
				},
				EntryexitId: 0,
			},
			expectedErr: status.Error(codes.InvalidArgument, "student entry exit id must be valid"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "empty entry date",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: nil,
					ExitDateTime:  timestamppb.New(testTime),
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "this field is required|date|time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry time is greater than exit time",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(testTime),
					ExitDateTime:  timestamppb.New(testTime.Add(-1 * time.Hour)),
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry time must be earlier than exit time|time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry date is ahead than exit date",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(testTime),
					ExitDateTime:  timestamppb.New(testTime.Add(-24 * time.Hour)),
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry date must be earlier than exit date|date"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry time is ahead than current time",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime.Add(time.Hour)),
					ExitDateTime:  nil,
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry time must not be a future time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry date is ahead than current date",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime.Add(24 * time.Hour)),
					ExitDateTime:  nil,
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry date must not be a future date"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "exit time is ahead than current time",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime),
					ExitDateTime:  timestamppb.New(currentTime.Add(time.Hour)),
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "exit time must not be a future time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "exit date is ahead than current date",
			ctx:  ctx,
			req: &eepb.UpdateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime),
					ExitDateTime:  timestamppb.New(currentTime.Add(24 * time.Hour)),
				},
				EntryexitId: 1,
			},
			expectedErr: status.Error(codes.InvalidArgument, "exit date must not be a future date"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name:        "failed to update record",
			ctx:         ctx,
			req:         entryExitRequestWithNotify(false),
			expectedErr: status.Error(codes.Internal, "closed pool"),
			setup: func(ctx context.Context) {
				testSetup.DefaultUpdateSetup(ctx, puddle.ErrClosedPool)
			},
		},
		{
			name:        "cannot retrieve student entry exit id err no rows",
			ctx:         ctx,
			req:         entryExitRequestWithNotify(false),
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			setup: func(ctx context.Context) {
				testSetup.DefaultUpdateSetup(ctx, pgx.ErrNoRows)
			},
		},
		{
			name:         "Happy case Update Entry and Exit Record with failed notification",
			ctx:          ctx,
			req:          entryExitRequestWithNotify(true),
			expectedResp: successResponseNotified(false),
			setup:        testSetup.UpdateWithFailedNotifSetup(nil, errors.New("publish error")),
		},
		{
			name:         "Happy case Update Entry Record with failed notification",
			ctx:          ctx,
			req:          entryRequestWithNotify(true),
			expectedResp: successResponseNotified(false),
			setup:        testSetup.UpdateWithFailedNotifSetup(nil, errors.New("publish error")),
		},
		{
			name:        "student user returns no rows result set",
			ctx:         ctx,
			req:         entryRequestWithNotify(true),
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			setup:       testSetup.UserSetup(nil, pgx.ErrNoRows),
		},
		{
			name:        "failed to retrieve user record",
			ctx:         ctx,
			req:         entryExitRequestWithNotify(false),
			expectedErr: status.Error(codes.Internal, "closed pool"),
			setup:       testSetup.UserSetup(nil, puddle.ErrClosedPool),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.UpdateEntryExit(testCase.ctx, testCase.req.(*eepb.UpdateEntryExitRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, s.DB, s.StudentEntryExitRecordsRepo, s.StudentRepo, s.StudentParentRepo, s.UserRepo)
		})
	}
}
