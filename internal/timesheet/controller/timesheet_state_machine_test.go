package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTimesheetController_ValidateTimesheetIDRequest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	testcases := []TestCase{
		{
			name:        "happy case valid timesheet id",
			ctx:         ctx,
			reqString:   "test-valid-id",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				// Do nothing no setup
			},
		},
		{
			name:        "invalid timesheet id",
			ctx:         ctx,
			reqString:   "",
			expectedErr: status.Error(codes.InvalidArgument, "timesheet id cannot be empty"),
			setup: func(ctx context.Context) {
				// Do nothing no setup
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := validateTimesheetIDRequest(testCase.reqString)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}

func TestTimesheetController_ValidateApproveConfirmTimesheetIDsRequest(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	testcases := []TestCase{
		{
			name:            "happy case valid single timesheet id request",
			ctx:             ctx,
			reqTimesheetIDs: []string{"1"},
			expectedErr:     nil,
			expectedResp:    []string{"1"},
			setup: func(ctx context.Context) {
				// Do nothing no setup
			},
		},
		{
			name:            "happy case valid multiple timesheet ids request",
			ctx:             ctx,
			reqTimesheetIDs: []string{"1", "2", "3"},
			expectedResp:    []string{"1", "2", "3"},
			expectedErr:     nil,
			setup: func(ctx context.Context) {
				// Do nothing no setup
			},
		},
		{
			name:            "invalid timesheet id",
			ctx:             ctx,
			reqTimesheetIDs: []string{},
			expectedResp:    []string{},
			expectedErr:     status.Error(codes.InvalidArgument, "timesheet ids cannot be empty"),
			setup: func(ctx context.Context) {
				// Do nothing no setup
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			err := validateApproveConfirmTimesheetIDsRequest(testCase.reqTimesheetIDs)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
