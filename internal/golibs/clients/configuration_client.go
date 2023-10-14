package clients

import (
	"context"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc"
)

type ConfigurationClient struct {
	configurationClient mpb.ConfigurationServiceClient
}

type ConfigurationClientInterface interface {
	GetConfigurations(ctx context.Context, req *mpb.GetConfigurationsRequest) (*mpb.GetConfigurationsResponse, error)
	GetConfigurationByKey(ctx context.Context, req *mpb.GetConfigurationByKeyRequest) (*mpb.GetConfigurationByKeyResponse, error)
}

func InitConfigurationClient(connect *grpc.ClientConn) *ConfigurationClient {
	configurationClient := mpb.NewConfigurationServiceClient(connect)
	return &ConfigurationClient{
		configurationClient: configurationClient,
	}
}

func (c *ConfigurationClient) GetConfigurations(ctx context.Context, req *mpb.GetConfigurationsRequest) (*mpb.GetConfigurationsResponse, error) {
	return c.configurationClient.GetConfigurations(ctx, req)
}

func (c *ConfigurationClient) GetConfigurationByKey(ctx context.Context, req *mpb.GetConfigurationByKeyRequest) (*mpb.GetConfigurationByKeyResponse, error) {
	return c.configurationClient.GetConfigurationByKey(ctx, req)
}
