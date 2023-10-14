package bob

import (
	"context"
	"errors"
	"strings"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) aUserGetClientVersion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewMasterDataServiceClient(s.Conn).GetClientVersion(s.signedCtx(ctx), &pb.GetClientVersionRequest{})
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) bobMustReturnsClientVersionFromConfig(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.GetClientVersionResponse)

	versions := strings.Split("com.manabie.student_manabie_app:1.1.0,com.manabie.studentManabieApp:1.1.0,com.manabie.liz:1.0.0", ",")
	for _, ver := range versions {
		parts := strings.Split(ver, ":")
		if resp.Versions[parts[0]] != parts[1] {
			return StepStateToContext(ctx, stepState), errors.New("result does not match")
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
