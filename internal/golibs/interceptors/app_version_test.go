package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewCheckAppVersion(t *testing.T) {
	t.Parallel()
	type args struct {
		cfgClientVersion string
		minVersionCheck  string
		methodIgnores    []string
	}

	testCases := []struct {
		name        string
		args        args
		expectedErr error
	}{
		{
			name:        "no client versions provide",
			args:        args{},
			expectedErr: errors.Wrap(errors.New("invalid version, must match pattern <pkg_name>:<required_version>"), "getAppVersion"),
		},
		{
			name:        "invalid client versions provide",
			args:        args{cfgClientVersion: "something-wrong"},
			expectedErr: errors.Wrap(errors.New("invalid version, must match pattern <pkg_name>:<required_version>"), "getAppVersion"),
		},
		{
			name: "valid client versions provide",
			args: args{
				cfgClientVersion: "com.manabie.student_manabie_app:1.1.0",
				minVersionCheck:  "1.0.0",
			},
			expectedErr: nil,
		},
		{
			name: "invalid version",
			args: args{
				cfgClientVersion: "com.manabie.student_manabie_app:1",
				minVersionCheck:  "1.0.0",
			},
			expectedErr: errors.Wrap(errors.New("invalid required version: invalid client version"), "getAppVersion"),
		},
		{
			name: "invalid minimum version",
			args: args{
				cfgClientVersion: "com.manabie.student_manabie_app:1.1.0",
				minVersionCheck:  "some-invalid-version",
			},
			expectedErr: errors.New("invalid client version"),
		},
	}

	for _, tc := range testCases {
		_, err := NewCheckAppVersion(tc.args.cfgClientVersion, tc.args.minVersionCheck, tc.args.methodIgnores)
		if tc.expectedErr != nil {
			assert.Equal(t, tc.expectedErr.Error(), err.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}

func TestUnaryServerCheckAppVersionInterceptor(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ignoreVersion := "0.0.8"

	testCases := []struct {
		name        string
		ctx         context.Context
		serverInfo  *grpc.UnaryServerInfo
		expectedErr error
	}{
		{
			name: "null metadata",
			ctx:  ctx,
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.Internal, "cannot check gRPC metadata"),
		},
		{
			name: "missing package name",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey: []string{"1.1.0"},
			}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing package name"),
		},
		{
			name: "success",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing client version"),
		},
		{
			name: "invalid package name",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"1.1.0"},
				packageNameKey: []string{"invalid-package-name"},
			}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid package name: invalid-package-name"),
		},
		{
			name: "invalid version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"invalid version"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid client version"),
		},
		{
			name: "force update",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"0.0.9"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.Aborted, "force update"),
		},
		{
			name: "ignore force update",
			ctx:  metadata.NewIncomingContext(ctx, metadata.MD{}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Ignore",
			},
			expectedErr: nil,
		},
		{
			name: "success",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"1.1.0"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: nil,
		},
		{
			name: "success less than min version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"0.1.7"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.UnaryServerInfo{
				Server:     nil,
				FullMethod: "Testing",
			},
			expectedErr: nil,
		},
	}

	checkClientVersion, _ := NewCheckAppVersion("com.manabie.student_app:1.1.0", ignoreVersion, []string{"Ignore"})

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			_, err := checkClientVersion.UnaryServerInterceptor(testcase.ctx, nil, testcase.serverInfo, func(ctx context.Context, req interface{}) (i interface{}, e error) {
				return nil, nil
			})
			assert.Equal(t, testcase.expectedErr, err)
		})
	}
}

func TestStreamServerCheckAppVersionInterceptor(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ignoreVersion := "0.0.8"

	testCases := []struct {
		name        string
		ctx         context.Context
		serverInfo  *grpc.StreamServerInfo
		expectedErr error
	}{
		{
			name: "null metadata",
			ctx:  ctx,
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.Internal, "cannot check gRPC metadata"),
		},
		{
			name: "missing package name",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey: []string{"1.0.0"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing package name"),
		},
		{
			name: "missing client version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing client version"),
		},
		{
			name: "invalid package name",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"1.0.0"},
				packageNameKey: []string{"invalid-package-name"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid package name: invalid-package-name"),
		},
		{
			name: "invalid version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"invalid version"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid client version"),
		},
		{
			name: "force update",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"0.0.9"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.Aborted, "force update"),
		},
		{
			name: "ignore force update",
			ctx:  metadata.NewIncomingContext(ctx, metadata.MD{}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Ignore",
			},
			expectedErr: nil,
		},
		{
			name: "success",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"1.1.0"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: nil,
		},
		{
			name: "success less than min version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"0.1.8"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: nil,
		},
	}

	checkClientVersion, _ := NewCheckAppVersion("com.manabie.student_app:1.1.0", ignoreVersion, []string{"Ignore"})

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			err := checkClientVersion.StreamServerInterceptor(nil, &mockServerStream{testcase.ctx}, testcase.serverInfo, func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			})

			assert.Equal(t, testcase.expectedErr, err)
		})
	}
}

func TestCheckForceUpdateApp(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	checkAppVersions := "com.manabie.student_app:1.1.0"
	testCases := []struct {
		name        string
		ctx         context.Context
		serverInfo  *grpc.StreamServerInfo
		expectedErr error
	}{
		{
			name: "null metadata",
			ctx:  ctx,
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.Internal, "cannot check gRPC metadata"),
		},
		{
			name: "success",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"1.1.0"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: nil,
		},
		{
			name: "force update",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"1.0.0"},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.Aborted, "force update"),
		},
		{
			name: "missing package name",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{"1.0.0"},
				packageNameKey: []string{},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing package name"),
		},
		{
			name: "missing version",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{},
				packageNameKey: []string{"com.manabie.student_app"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "missing client version"),
		},
		{
			name: "invalid package name",
			ctx: metadata.NewIncomingContext(ctx, metadata.MD{
				versionKey:     []string{},
				packageNameKey: []string{"com.manabie.liz"},
			}),
			serverInfo: &grpc.StreamServerInfo{
				FullMethod: "Testing",
			},
			expectedErr: status.Error(codes.InvalidArgument, "invalid package name: com.manabie.liz"),
		},
	}

	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			err := CheckForceUpdateApp(testcase.ctx, checkAppVersions)

			assert.Equal(t, testcase.expectedErr, err)
		})
	}

}

func TestCompareVersion(t *testing.T) {
	testCases := []struct {
		name            string
		requiredVersion string
		appVersion      string
		expectedErr     error
	}{
		{
			name:            "force update",
			requiredVersion: "1.0.0",
			appVersion:      "0.2.20220831",
			expectedErr:     status.Error(codes.Aborted, "force update"),
		},
		{
			name:            "success",
			requiredVersion: "1.2.5678",
			appVersion:      "1.3.432",
			expectedErr:     nil,
		},
		{
			name:            "minor version force update",
			requiredVersion: "1.5.1234",
			appVersion:      "1.4.1234",
			expectedErr:     status.Error(codes.Aborted, "force update"),
		},
		{
			name:            "major version success",
			requiredVersion: "5.123.4325",
			appVersion:      "12.51.1235",
			expectedErr:     nil,
		},
		{
			name:            "spec test case 1",
			requiredVersion: "1.0.20220923020330",
			appVersion:      "2.0.20220923020330",
			expectedErr:     nil,
		},
		{
			name:            "spec test case 2",
			requiredVersion: "1.0.20220923020330",
			appVersion:      "1.0.20220823020330",
			expectedErr:     status.Error(codes.Aborted, "force update"),
		},
		{
			name:            "spec test case 3",
			requiredVersion: "1.0.20220923020330",
			appVersion:      "1.0.20221023020330",
			expectedErr:     nil,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			err := compareVersion(testcase.requiredVersion, testcase.appVersion)

			assert.Equal(t, testcase.expectedErr, err)
		})
	}
}

type mockServerStream struct {
	ctx context.Context
}

func (*mockServerStream) SetHeader(metadata.MD) error {
	return nil
}

func (*mockServerStream) SendHeader(metadata.MD) error {
	return nil
}

func (*mockServerStream) SetTrailer(metadata.MD) {
}

func (rcv *mockServerStream) Context() context.Context {
	return rcv.ctx
}

func (*mockServerStream) SendMsg(m interface{}) error {
	return nil
}

func (*mockServerStream) RecvMsg(m interface{}) error {
	return nil
}
