package paymentsvc

import (
	"context"
)

// nolint:unused,structcheck
type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}
