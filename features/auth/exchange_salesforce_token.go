package auth

import (
	"context"
	"fmt"

	apb "github.com/manabie-com/backend/pkg/manabuf/auth/v1"
)

func (s *suite) userExchangesSalesforceToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := apb.NewAuthServiceClient(s.AuthConn).
		ExchangeSalesforceToken(ctx, &apb.ExchangeSalesforceTokenRequest{})
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userExchangesSalesforceTokenSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	resp := stepState.Response.(*apb.ExchangeSalesforceTokenResponse)
	if resp.Token == "" {
		return ctx, fmt.Errorf("exchange salesforce token failed")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userCanNotExchangesSalesforceTokenWithStatusCode(ctx context.Context, code string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr == nil {
		return ctx, fmt.Errorf("expected error is not nil")
	}

	if err := compareStatusCode(stepState.ResponseErr, code); err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}
