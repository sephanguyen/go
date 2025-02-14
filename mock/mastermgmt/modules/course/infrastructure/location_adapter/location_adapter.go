// Code generated by mockgen. DO NOT EDIT.
package mock_locationadapter

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type MockLocationAdapter struct {
	mock.Mock
}

func (r *MockLocationAdapter) GetLocationsByLocationIDs(arg1 context.Context, arg2 database.Ext, arg3 []string) ([]string, error) {
	args := r.Called(arg1, arg2, arg3)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
