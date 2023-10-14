package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"google.golang.org/grpc/status"
)

func (s *suite) theMessageIsEchoed(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	req := stepState.Request.(*pb.EchoRequest)
	resp := stepState.Response.(*pb.EchoResponse)
	if req.Message != resp.Message {
		return ctx, fmt.Errorf("expecting response message %s, got %s", req.Message, resp.Message)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) echoAMessage(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewEchoServiceClient(s.PaymentConn).Echo(contextWithToken(ctx), stepState.Request.(*pb.EchoRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anEchoMessage(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	uniqueID := idutil.ULIDNow()
	stepState.Request = &pb.EchoRequest{
		Message: "message random " + uniqueID,
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) receivesStatusCode(ctx context.Context, expectedCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	if stt.Code().String() != expectedCode {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", expectedCode, stt.Code().String(), stt.Message())
	}

	return StepStateToContext(ctx, stepState), nil
}
