// Code generated by mockgen. DO NOT EDIT.
package mock_external

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockCerebryRepo struct {
	mock.Mock
}

func (r *MockCerebryRepo) GetUserToken(arg1 context.Context, arg2 string) (string, error) {
	args := r.Called(arg1, arg2)
	return args.Get(0).(string), args.Error(1)
}
