package controller

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	external_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/configuration/infrastructure/repo"
	mock_external_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/external_configuration/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestConfigurationService_GetConfigurations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	configRepo := &mock_repo.MockConfigRepo{}
	externalConfigRepo := &mock_external_repo.MockExternalConfigRepo{}
	existing := make([]*domain.InternalConfiguration, 10)
	for i := 0; i < 10; i++ {
		id := idutil.ULIDNow()
		existing[i] = &domain.InternalConfiguration{
			ID:              id,
			ConfigKey:       "z-key-" + strconv.Itoa(i) + id,
			ConfigValue:     "value-" + strconv.Itoa(i) + id,
			ConfigValueType: "string",
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	}
	s := NewConfigurationService(db, configRepo, externalConfigRepo)

	tc := []TestCase{
		{
			name: "config found",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationsRequest{Keyword: "key-2", Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: nil,
			expectedResp: &mpb.GetConfigurationsResponse{
				Items: []*mpb.Configuration{
					{
						Id:              existing[2].ID,
						ConfigKey:       existing[2].ConfigKey,
						ConfigValue:     existing[2].ConfigValue,
						ConfigValueType: existing[2].ConfigValueType,
						CreatedAt:       existing[2].CreatedAt.String(),
						UpdatedAt:       existing[2].UpdatedAt.String(),
					},
				},
				NextPage: &cpb.Paging{
					Limit: uint32(10),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On("SearchWithKey", ctx, db, domain.ConfigSearchArgs{
					Keyword: "key-2", Limit: int64(10), Offset: int64(0),
				}).
					Return([]*domain.InternalConfiguration{
						existing[2],
					}, nil).Once()
			},
		},
		{
			name: "config not found",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationsRequest{Keyword: "not-found-key-2", Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: nil,
			expectedResp: &mpb.GetConfigurationsResponse{
				Items: []*mpb.Configuration{},
				NextPage: &cpb.Paging{
					Limit: uint32(10),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On("SearchWithKey", ctx, db, domain.ConfigSearchArgs{
					Keyword: "not-found-key-2", Limit: int64(10), Offset: int64(0),
				}).
					Return([]*domain.InternalConfiguration{}, nil).Once()

				externalConfigRepo.On("SearchWithKey", ctx, db, external_domain.ExternalConfigSearchArgs{
					Keyword: "not-found-key-2", Limit: int64(10), Offset: int64(0),
				}).
					Return([]*external_domain.ExternalConfiguration{}, nil).Once()
			},
		},
		{
			name: "internal err",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationsRequest{Keyword: "some-key-2", Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr:  status.Error(codes.Internal, "internal err"),
			expectedResp: &mpb.GetConfigurationsResponse{},
			setup: func(ctx context.Context) {
				configRepo.On("SearchWithKey", ctx, db, domain.ConfigSearchArgs{
					Keyword: "some-key-2", Limit: int64(10), Offset: int64(0),
				}).
					Return([]*domain.InternalConfiguration{}, errors.New("internal err")).Once()
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetConfigurationsRequest)
			resp, err := s.GetConfigurations(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, db, configRepo)
		})
	}
}

func TestConfigurationService_GetConfigurationByKey(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	configRepo := &mock_repo.MockConfigRepo{}
	externalConfigRepo := &mock_external_repo.MockExternalConfigRepo{}
	existing := make([]*domain.InternalConfiguration, 10)
	for i := 0; i < 10; i++ {
		id := idutil.ULIDNow()
		existing[i] = &domain.InternalConfiguration{
			ID:          id,
			ConfigKey:   "z-key-" + strconv.Itoa(i) + id,
			ConfigValue: "value-" + strconv.Itoa(i) + id,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	s := NewConfigurationService(db, configRepo, externalConfigRepo)

	tc := []TestCase{
		{
			name:        "config found",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &mpb.GetConfigurationByKeyRequest{Key: existing[2].ConfigKey},
			expectedErr: nil,
			expectedResp: &mpb.GetConfigurationByKeyResponse{
				Configuration: &mpb.Configuration{
					Id:          existing[2].ID,
					ConfigKey:   existing[2].ConfigKey,
					ConfigValue: existing[2].ConfigValue,
					CreatedAt:   existing[2].CreatedAt.String(),
					UpdatedAt:   existing[2].UpdatedAt.String(),
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, db, existing[2].ConfigKey).
					Return(&domain.InternalConfiguration{
						ID:          existing[2].ID,
						ConfigKey:   existing[2].ConfigKey,
						ConfigValue: existing[2].ConfigValue,
						CreatedAt:   existing[2].CreatedAt,
						UpdatedAt:   existing[2].UpdatedAt,
					}, nil).Times(1)
			},
		},
		{
			name:         "config not found",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.GetConfigurationByKeyRequest{Key: "not-found-key-2"},
			expectedErr:  status.Error(codes.NotFound, pgx.ErrNoRows.Error()),
			expectedResp: &mpb.GetConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, db, "not-found-key-2").
					Return(nil, pgx.ErrNoRows).Times(1)
			},
		},
		{
			name:         "empty key",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.GetConfigurationByKeyRequest{Key: "  "},
			expectedErr:  status.Error(codes.FailedPrecondition, "configuration key cannot be empty"),
			expectedResp: &mpb.GetConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "internal err",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.GetConfigurationByKeyRequest{Key: "some-key"},
			expectedErr:  status.Error(codes.Internal, "internal err"),
			expectedResp: &mpb.GetConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, db, "some-key").
					Return(nil, errors.New("internal err")).Times(1)
			},
		},
	}

	for _, testCase := range tc {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetConfigurationByKeyRequest)
			resp, err := s.GetConfigurationByKey(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedResp, resp)
			}
			mock.AssertExpectationsForObjects(t, db, configRepo)
		})
	}
}
