package bob

import (
	"context"
	"fmt"
	"time"

	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"google.golang.org/grpc/metadata"
)

func (s *suite) verifyAppVersionRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("pkg", "com.manabie.learner", "version", "2.0.0"))

	stepState.Request = &bpb.VerifyAppVersionRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCheckAppVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = bpb.NewInternalReaderServiceClient(s.Conn).VerifyAppVersion(ctx, stepState.Request.(*bpb.VerifyAppVersionRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) verifyAppVersionRequestMissing(ctx context.Context, missingData string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	version := "1.0.0"
	packageName := "com.manabie.learner"

	switch missingData {
	case "packageName":
		packageName = ""
	case "version":
		version = ""
	}

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("pkg", packageName, "version", version))
	stepState.Request = &bpb.VerifyAppVersionRequest{}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) verifyAppVersionRequestWith(ctx context.Context, appVersion string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("pkg", "com.manabie.learner", "version", appVersion))

	stepState.Request = &bpb.VerifyAppVersionRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userVerifyAppVersionReceiveForceUpdateRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return ctx, fmt.Errorf("expected response has err but actual is nil")
	}

	return ctx, nil
}
