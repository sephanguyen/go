package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestVerifyAppVersion(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	packageName := "com.manabie.liz"
	version := "1.1.0"
	checkClientVersions := []string{
		fmt.Sprintf("%s:%s", packageName, version),
	}

	testCases := []struct {
		name        string
		ctx         context.Context
		serverInfo  *grpc.StreamServerInfo
		req         interface{}
		expectedErr error
	}{
		{
			name: "null metadata",
			ctx:  ctx,
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			req:         &bpb.VerifyAppVersionRequest{},
			expectedErr: status.Error(codes.Internal, "cannot check gRPC metadata"),
		},
		{
			name: "success",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{version},
				"pkg":     []string{packageName},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			req:         &bpb.VerifyAppVersionRequest{},
			expectedErr: nil,
		},
		{
			name: "missing package name",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{version},
				"pkg":     []string{},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			req:         &bpb.VerifyAppVersionRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "missing package name"),
		},
		{
			name: "missing version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{},
				"pkg":     []string{packageName},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			req:         &bpb.VerifyAppVersionRequest{},
			expectedErr: status.Error(codes.InvalidArgument, "missing client version"),
		},
		{
			name: "force update",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				"version": []string{"1.0.0"},
				"pkg":     []string{packageName},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			req:         &bpb.VerifyAppVersionRequest{},
			expectedErr: status.Error(codes.Aborted, "force update"),
		},
	}

	svc := &InternalReaderService{
		CheckClientVersions: checkClientVersions,
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := testCase.req.(*bpb.VerifyAppVersionRequest)
			_, err := svc.VerifyAppVersion(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
