// Code generated by mockgen. DO NOT EDIT.
package mock_clients

import (
	"context"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/mock"
)

type MockConfigurationClient struct {
	mock.Mock
}

func (m *MockConfigurationClient) GetConfigurationByKey(arg1 context.Context, arg2 *mpb.GetConfigurationByKeyRequest) (*mpb.GetConfigurationByKeyResponse, error) {
	args := m.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mpb.GetConfigurationByKeyResponse), args.Error(1)
}

func (m *MockConfigurationClient) GetConfigurations(arg1 context.Context, arg2 *mpb.GetConfigurationsRequest) (*mpb.GetConfigurationsResponse, error) {
	args := m.Called(arg1, arg2)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mpb.GetConfigurationsResponse), args.Error(1)
}
