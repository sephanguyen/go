package controller

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/application/queries"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/infrastructure"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const PageLimit int64 = 10000

type ConfigurationService struct {
	GetConfigQueryHandler queries.GetConfigurationQueryHandler
}

func NewConfigurationService(
	db database.Ext,
	configRepo infrastructure.ConfigRepo,
	externalConfigRepo infrastructure.ExternalConfigRepo,
) *ConfigurationService {
	return &ConfigurationService{
		GetConfigQueryHandler: queries.GetConfigurationQueryHandler{
			DB:                 db,
			ConfigRepo:         configRepo,
			ExternalConfigRepo: externalConfigRepo,
		},
	}
}

func (c *ConfigurationService) GetConfigurations(ctx context.Context, req *mpb.GetConfigurationsRequest) (resp *mpb.GetConfigurationsResponse, err error) {
	payload := queries.GetConfigurations{
		SearchOption: domain.ConfigSearchArgs{
			Offset:  0,
			Limit:   PageLimit,
			Keyword: req.Keyword,
		},
	}

	if req.Paging != nil && req.Paging.Limit != 0 {
		payload.SearchOption.Limit = int64(req.Paging.Limit)
		payload.SearchOption.Offset = req.Paging.GetOffsetInteger()
	}
	cfs, err := c.GetConfigQueryHandler.SearchWithKey(ctx, payload)
	if err != nil {
		return &mpb.GetConfigurationsResponse{}, status.Error(codes.Internal, err.Error())
	}
	pcfs := make([]*mpb.Configuration, len(cfs))
	for i, v := range cfs {
		pcfs[i] = &mpb.Configuration{
			Id:              v.ID,
			ConfigKey:       v.ConfigKey,
			ConfigValue:     v.ConfigValue,
			ConfigValueType: v.ConfigValueType,
			CreatedAt:       v.CreatedAt.String(),
			UpdatedAt:       v.UpdatedAt.String(),
		}
	}
	return &mpb.GetConfigurationsResponse{
		Items: pcfs,
		NextPage: &cpb.Paging{
			Limit: uint32(payload.SearchOption.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: payload.SearchOption.Offset + payload.SearchOption.Limit,
			},
		},
	}, nil
}

func (c *ConfigurationService) GetConfigurationByKey(ctx context.Context, req *mpb.GetConfigurationByKeyRequest) (resp *mpb.GetConfigurationByKeyResponse, err error) {
	if strings.TrimSpace(req.Key) == "" {
		return &mpb.GetConfigurationByKeyResponse{}, status.Error(codes.FailedPrecondition, "configuration key cannot be empty")
	}
	payload := queries.GetConfigurationByKey{
		Key: req.Key,
	}
	cf, err := c.GetConfigQueryHandler.GetByKey(ctx, payload)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &mpb.GetConfigurationByKeyResponse{}, status.Error(codes.NotFound, err.Error())
		}
		return &mpb.GetConfigurationByKeyResponse{}, status.Error(codes.Internal, err.Error())
	}
	return &mpb.GetConfigurationByKeyResponse{
		Configuration: &mpb.Configuration{
			Id:          cf.ID,
			ConfigKey:   cf.ConfigKey,
			ConfigValue: cf.ConfigValue,
			CreatedAt:   cf.CreatedAt.String(),
			UpdatedAt:   cf.UpdatedAt.String(),
		},
	}, nil
}
