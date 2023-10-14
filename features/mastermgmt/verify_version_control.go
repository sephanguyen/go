package mastermgmt

import (
	"context"
	"fmt"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc/metadata"
)

func (s *suite) aInvalidVersionRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &mpb.VerifyAppVersionRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userVerifyVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*mpb.VerifyAppVersionRequest)
	stepState.Response, stepState.ResponseErr = mpb.NewVersionControlReaderServiceClient(s.MasterMgmtConn).VerifyAppVersion(ctx, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aRequestWithLowerVersion(ctx context.Context, version string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.student_manabie_app", "version", version)
	stepState.Request = &mpb.VerifyAppVersionRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aRequestWithValidVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.student_manabie_app", "version", "1.5.0")
	stepState.Request = &mpb.VerifyAppVersionRequest{}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnFalseInMessage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp := stepState.Response.(*mpb.VerifyAppVersionResponse)
	if resp.IsValid {
		return nil, fmt.Errorf("expected false, got true for verify version")
	}

	return StepStateToContext(ctx, stepState), nil
}
