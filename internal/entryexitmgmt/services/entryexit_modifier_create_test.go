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

func TestEntryExitModifierService_CreateEntryExit(t *testing.T) {
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

	entryRequestWithNotify := func(notify bool) *eepb.CreateEntryExitRequest {
		return &eepb.CreateEntryExitRequest{
			EntryExitPayload: &eepb.EntryExitPayload{
				StudentId:     "student-exist",
				EntryDateTime: timestamppb.New(testTime),
				ExitDateTime:  nil,
				NotifyParents: notify,
			},
		}
	}
	entryExitRequestWithNotify := func(notify bool) *eepb.CreateEntryExitRequest {
		return &eepb.CreateEntryExitRequest{
			EntryExitPayload: &eepb.EntryExitPayload{
				StudentId:     "student-exist",
				EntryDateTime: timestamppb.New(testTime.Add(-1 * time.Hour)),
				ExitDateTime:  timestamppb.New(testTime),
				NotifyParents: notify,
			},
		}
	}
	successResponseNotified := func(notify bool) *eepb.CreateEntryExitResponse {
		return &eepb.CreateEntryExitResponse{
			Successful:     true,
			Message:        "You have added a new record successfully!",
			ParentNotified: notify,
		}
	}

	testCases := []TestCase{
		{
			name:         "Happy case Only Entry Record",
			ctx:          ctx,
			req:          entryRequestWithNotify(false),
			expectedResp: successResponseNotified(false),
			setup:        testSetup.CreateSetup(),
		},
		{
			name:         "Happy case both Entry and Exit Record",
			ctx:          ctx,
			req:          entryExitRequestWithNotify(false),
			expectedResp: successResponseNotified(false),
			setup:        testSetup.CreateSetup(),
		},
		{
			name:         "Happy case Only Entry Record with notification",
			ctx:          ctx,
			req:          entryRequestWithNotify(true),
			expectedResp: successResponseNotified(true),
			setup:        testSetup.CreateWithNotifSetup(),
		},
		{
			name:         "Happy case both Entry and Exit Record with notification",
			ctx:          ctx,
			req:          entryExitRequestWithNotify(true),
			expectedResp: successResponseNotified(true),
			setup:        testSetup.CreateWithNotifSetup(),
		},
		{
			name:        "failed to get parent ids closed db pool",
			ctx:         ctx,
			req:         entryExitRequestWithNotify(true),
			expectedErr: status.Error(codes.Internal, "closed pool"),
			setup:       testSetup.CreateWithGetParentIDSetup(testSetup.MockIDs, (puddle.ErrClosedPool)),
		},
		{
			name:        "failed to get parent ids err no rows",
			ctx:         ctx,
			req:         entryExitRequestWithNotify(true),
			expectedErr: status.Error(codes.Internal, "no rows in result set"),
			setup:       testSetup.CreateWithGetParentIDSetup(make([]string, 0), pgx.ErrNoRows),
		},
		{
			name: "student id is empty",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId: "",
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student id cannot be empty"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "empty entry date",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: nil,
					ExitDateTime:  timestamppb.New(testTime),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "this field is required|date|time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry time is greater than exit time",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(testTime),
					ExitDateTime:  timestamppb.New(testTime.Add(-1 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry time must be earlier than exit time|time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry date is ahead than exit date",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(testTime),
					ExitDateTime:  timestamppb.New(testTime.Add(-24 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry date must be earlier than exit date|date"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry time is ahead than current time",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime.Add(time.Hour)),
					ExitDateTime:  nil,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry time must not be a future time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "entry date is ahead than current date",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime.UTC().Add(24 * time.Hour)),
					ExitDateTime:  nil,
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "entry date must not be a future date"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "exit time is ahead than current time",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime.UTC()),
					ExitDateTime:  timestamppb.New(currentTime.UTC().Add(time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "exit time must not be a future time"),
			setup:       testSetup.EmptySetup(),
		},
		{
			name: "exit date is ahead than current date",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "some-id",
					EntryDateTime: timestamppb.New(currentTime.UTC()),
					ExitDateTime:  timestamppb.New(currentTime.UTC().Add(24 * time.Hour)),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "exit date must not be a future date"),
			setup:       testSetup.EmptySetup(),
		},
		// student not existing in database
		{
			name: "cannot create if student id not existing in mockDB",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "not-exist-id",
					EntryDateTime: timestamppb.New(testTime.Add(-1 * time.Hour)),
					ExitDateTime:  timestamppb.New(testTime),
				},
			},
			expectedErr: status.Error(codes.InvalidArgument, "student id does not exist"),
			setup:       testSetup.StudentNotExistSetup(),
		},
		{
			name: "failed to create record",
			ctx:  ctx,
			req: &eepb.CreateEntryExitRequest{
				EntryExitPayload: &eepb.EntryExitPayload{
					StudentId:     "not-exist-id",
					EntryDateTime: timestamppb.New(testTime.Add(-1 * time.Hour)),
					ExitDateTime:  timestamppb.New(testTime),
				},
			},
			expectedErr: status.Error(codes.Internal, "closed pool"),
			setup: func(ctx context.Context) {
				testSetup.DefaultCreateSetup(ctx, puddle.ErrClosedPool)
			},
		},
		{
			name:         "Happy case Only Entry Record with failed notification",
			ctx:          ctx,
			req:          entryRequestWithNotify(true),
			expectedErr:  nil,
			expectedResp: successResponseNotified(false),
			setup:        testSetup.CreateWithFailedNotifSetup(nil, errors.New("publish error")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			resp, err := s.CreateEntryExit(testCase.ctx, testCase.req.(*eepb.CreateEntryExitRequest))
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
			mock.AssertExpectationsForObjects(t, s.DB, s.StudentEntryExitRecordsRepo, s.StudentRepo, s.StudentParentRepo)
		})
	}

}
