package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb_ms "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	name        string
	ctx         context.Context
	req         interface{}
	expectedErr error
	setup       func(ctx context.Context)
	expectedRsp interface{}
}

func TestVerifyAppVersion(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s := VersionControlService{
		JoinClientVersions: "com.manabie.student_manabie_app:1.5.0",
	}

	testCases := []TestCase{
		{
			name: "force update",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{"1.0.0"},
				"pkg":     []string{"com.manabie.student_manabie_app"},
			}),
			req: &pb_ms.VerifyAppVersionRequest{},
			expectedRsp: &pb_ms.VerifyAppVersionResponse{
				IsValid: false,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
		{
			name: "force update major version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{"0.5.0"},
				"pkg":     []string{"com.manabie.student_manabie_app"},
			}),
			req: &pb_ms.VerifyAppVersionRequest{},
			expectedRsp: &pb_ms.VerifyAppVersionResponse{
				IsValid: false,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
		{
			name: "invalid package",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{"1.0.0"},
				"pkg":     []string{"com.manabie.student"},
			}),
			req: &pb_ms.VerifyAppVersionRequest{},
			expectedRsp: &pb_ms.VerifyAppVersionResponse{
				IsValid: false,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid package name: com.manabie.student"),
		},
		{
			name: "happy case",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{"1.5.0"},
				"pkg":     []string{"com.manabie.student_manabie_app"},
			}),
			req: &pb_ms.VerifyAppVersionRequest{},
			expectedRsp: &pb_ms.VerifyAppVersionResponse{
				IsValid: true,
			},
			setup: func(ctx context.Context) {
			},
			expectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			testCase.setup(testCase.ctx)

			rsp, err := s.VerifyAppVersion(testCase.ctx, testCase.req.(*pb_ms.VerifyAppVersionRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
				assert.Equal(t, testCase.expectedRsp.(*pb_ms.VerifyAppVersionResponse), rsp)
			}
		})
	}
}
