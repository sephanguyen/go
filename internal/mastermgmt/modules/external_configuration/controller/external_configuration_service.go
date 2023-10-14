package controller

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/application/commands"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/infrastructure"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const PageLimit int64 = 100

type ExternalConfigurationService struct {
	GetExternalConfigQueryHandler      queries.GetExternalConfigurationQueryHandler
	CreateExternalConfigurationHandler commands.CreateExternalConfigurationHandler
}

func NewExternalConfigurationService(
	db database.Ext,
	configRepo infrastructure.ExternalConfigRepo,
) *ExternalConfigurationService {
	return &ExternalConfigurationService{
		GetExternalConfigQueryHandler: queries.GetExternalConfigurationQueryHandler{
			DB:         db,
			ConfigRepo: configRepo,
		},
		CreateExternalConfigurationHandler: commands.CreateExternalConfigurationHandler{
			DB:         db,
			ConfigRepo: configRepo,
		},
	}
}

func (e *ExternalConfigurationService) GetExternalConfigurations(ctx context.Context, req *mpb.GetExternalConfigurationsRequest) (resp *mpb.GetExternalConfigurationsResponse, err error) {
	payload := queries.GetExternalConfigurations{
		SearchOption: domain.ExternalConfigSearchArgs{
			Offset:  0,
			Limit:   PageLimit,
			Keyword: req.Keyword,
		},
	}

	if req.Paging != nil && req.Paging.Limit != 0 {
		payload.SearchOption.Limit = int64(req.Paging.Limit)
		payload.SearchOption.Offset = req.Paging.GetOffsetInteger()
	}
	cfs, err := e.GetExternalConfigQueryHandler.SearchWithKey(ctx, payload)
	if err != nil {
		return &mpb.GetExternalConfigurationsResponse{}, status.Error(codes.Internal, err.Error())
	}
	pcfs := make([]*mpb.ExternalConfiguration, len(cfs))
	for i, v := range cfs {
		pcfs[i] = &mpb.ExternalConfiguration{
			Id:          v.ID,
			ConfigKey:   v.ConfigKey,
			ConfigValue: v.ConfigValue,
			CreatedAt:   v.CreatedAt.String(),
			UpdatedAt:   v.UpdatedAt.String(),
		}
	}
	return &mpb.GetExternalConfigurationsResponse{
		Items: pcfs,
		NextPage: &cpb.Paging{
			Limit: uint32(payload.SearchOption.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: payload.SearchOption.Offset + payload.SearchOption.Limit,
			},
		},
	}, nil
}

func (e *ExternalConfigurationService) GetExternalConfigurationByKey(ctx context.Context, req *mpb.GetExternalConfigurationByKeyRequest) (resp *mpb.GetExternalConfigurationByKeyResponse, err error) {
	if strings.TrimSpace(req.Key) == "" {
		return &mpb.GetExternalConfigurationByKeyResponse{}, status.Error(codes.FailedPrecondition, "configuration key cannot be empty")
	}
	payload := queries.GetExternalConfigurationByKey{
		Key: req.Key,
	}
	cf, err := e.GetExternalConfigQueryHandler.GetByKey(ctx, payload)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &mpb.GetExternalConfigurationByKeyResponse{}, status.Error(codes.NotFound, err.Error())
		}
		return &mpb.GetExternalConfigurationByKeyResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &mpb.GetExternalConfigurationByKeyResponse{
		Configuration: &mpb.ExternalConfiguration{
			Id:          cf.ID,
			ConfigKey:   cf.ConfigKey,
			ConfigValue: cf.ConfigValue,
			CreatedAt:   cf.CreatedAt.String(),
			UpdatedAt:   cf.UpdatedAt.String(),
		},
	}, nil
}

func (e *ExternalConfigurationService) CreateMultiConfigurations(ctx context.Context, req *mpb.CreateMultiConfigurationsRequest) (*mpb.CreateMultiConfigurationsResponse, error) {
	now := time.Now()
	payload := sliceutils.Map(req.ExternalConfigurations, func(c *mpb.CreateMultiConfigurationsRequest_ExternalConfiguration) *domain.ExternalConfiguration {
		data := &domain.ExternalConfiguration{}

		data.ConfigKey = c.Key
		data.ConfigValue = c.Value
		data.ConfigValueType = c.ValueType
		data.ID = idutil.ULIDNow()
		data.CreatedAt = now
		data.UpdatedAt = now
		return data
	})
	err := e.CreateExternalConfigurationHandler.CreateMultiConfigurations(ctx, payload)
	if err != nil {
		return &mpb.CreateMultiConfigurationsResponse{Successful: false}, status.Error(codes.Internal, err.Error())
	}
	return &mpb.CreateMultiConfigurationsResponse{Successful: true}, nil
}

// Deprecated: please use GetConfigurationByKeysAndLocationsV2 which support optional locations
func (e *ExternalConfigurationService) GetConfigurationByKeysAndLocations(ctx context.Context, req *mpb.GetConfigurationByKeysAndLocationsRequest) (*mpb.GetConfigurationByKeysAndLocationsResponse, error) {
	if len(req.GetKeys()) == 0 || len(req.GetLocationsIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "keys and location ids are required fields in request")
	}
	cf, err := e.GetExternalConfigQueryHandler.GetLocationConfigByKeysAndLocations(ctx, req.GetKeys(), req.GetLocationsIds())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*mpb.LocationConfiguration, 0, len(cf))
	for _, element := range cf {
		res = append(res, element.ToLocationConfigurationGRPCMessage())
	}

	return &mpb.GetConfigurationByKeysAndLocationsResponse{
		Configurations: res,
	}, nil
}

func (e *ExternalConfigurationService) GetConfigurationByKeysAndLocationsV2(ctx context.Context, req *mpb.GetConfigurationByKeysAndLocationsV2Request) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
	if len(req.GetKeys()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "configuration key is required")
	}

	cf, err := e.GetExternalConfigQueryHandler.GetLocationConfigByKeys(ctx, req.GetKeys(), req.GetLocationIds())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*mpb.LocationConfiguration, 0, len(cf))
	for _, element := range cf {
		res = append(res, element.ToLocationConfigurationGRPCMessage())
	}

	return &mpb.GetConfigurationByKeysAndLocationsV2Response{
		Configurations: res,
	}, nil
}
