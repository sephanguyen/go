package utils

import (
	"context"
)

type TestCase struct {
	Name                string
	Ctx                 context.Context
	Req                 interface{}
	ExpectedResp        interface{}
	ExpectedErr         error
	ExpectErrorMessages interface{}
	Setup               func(ctx context.Context)
}
