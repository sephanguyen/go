package mastermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/timesheet/domain/common"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/grpc"
)

type MasterConfigurationServiceImpl struct {
	MasterMgmtConfigurationServiceClient interface {
		GetConfigurationByKey(context.Context, *mpb.GetConfigurationByKeyRequest, ...grpc.CallOption) (*mpb.GetConfigurationByKeyResponse, error)
	}

	MasterMgmtInternalServiceClient interface {
		GetConfigurations(context.Context, *mpb.GetConfigurationsRequest, ...grpc.CallOption) (*mpb.GetConfigurationsResponse, error)
	}
}

func (s *MasterConfigurationServiceImpl) CheckPartnerTimesheetServiceIsOn(ctx context.Context) (bool, error) {

	getConfigurationByKeyRequest := &mpb.GetConfigurationByKeyRequest{
		Key: constant.TimesheetServiceConfigurationKey,
	}

	res, err := s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey(common.SignCtx(ctx), getConfigurationByKeyRequest)

	if err != nil {
		return false, fmt.Errorf("s.MasterMgmtConfigurationServiceClient.GetConfigurationByKey: %s", err)
	}

	configValue := res.GetConfiguration().GetConfigValue()
	configValueOn := constant.ConfigSettingStatus.String(constant.On)

	if configValue == configValueOn {
		return true, nil
	}

	return false, nil

}

func (s *MasterConfigurationServiceImpl) CheckPartnerTimesheetServiceIsOnWithoutToken(ctx context.Context) (bool, error) {
	resourcePath, err := interceptors.ResourcePathFromContext(ctx)
	if err != nil {
		return false, fmt.Errorf("s.MastermgmtInternalServiceImpl.GetConfigurations: %s", err)
	}
	getConfigurationsRequest := &mpb.GetConfigurationsRequest{
		Paging:         &cpb.Paging{},
		Keyword:        constant.TimesheetServiceConfigurationKey,
		OrganizationId: resourcePath,
	}

	res, err := s.MasterMgmtInternalServiceClient.GetConfigurations(common.SignCtx(ctx), getConfigurationsRequest)

	if err != nil {
		return false, fmt.Errorf("s.MastermgmtInternalServiceImpl.GetConfigurations: %s", err)
	}

	configurationResponse := res.GetItems()[0]
	configValue := configurationResponse.GetConfigValue()
	configValueOn := constant.ConfigSettingStatus.String(constant.On)

	if configValue == configValueOn {
		return true, nil
	}

	return false, nil

}
