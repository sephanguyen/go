package controller

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/external_configuration/infrastructure/repo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestConfigurationService_GetExternalConfigurations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	configRepo := &mock_repo.MockExternalConfigRepo{}
	existing := make([]*domain.ExternalConfiguration, 10)
	for i := 0; i < 10; i++ {
		id := idutil.ULIDNow()
		existing[i] = &domain.ExternalConfiguration{
			ID:          id,
			ConfigKey:   "z-key-" + strconv.Itoa(i) + id,
			ConfigValue: "value-" + strconv.Itoa(i) + id,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	s := NewExternalConfigurationService(db, configRepo)

	tc := []TestCase{
		{
			name: "config found",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetExternalConfigurationsRequest{Keyword: "key-2", Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: nil,
			expectedResp: &mpb.GetExternalConfigurationsResponse{
				Items: []*mpb.ExternalConfiguration{
					{
						Id:          existing[2].ID,
						ConfigKey:   existing[2].ConfigKey,
						ConfigValue: existing[2].ConfigValue,
						CreatedAt:   existing[2].CreatedAt.String(),
						UpdatedAt:   existing[2].UpdatedAt.String(),
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
				configRepo.On("SearchWithKey", ctx, db, domain.ExternalConfigSearchArgs{
					Keyword: "key-2", Limit: int64(10), Offset: int64(0),
				}).
					Return([]*domain.ExternalConfiguration{
						existing[2],
					}, nil).Once()
			},
		},
		{
			name: "config not found",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetExternalConfigurationsRequest{Keyword: "not-found-key-2", Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr: nil,
			expectedResp: &mpb.GetExternalConfigurationsResponse{
				Items: []*mpb.ExternalConfiguration{},
				NextPage: &cpb.Paging{
					Limit: uint32(10),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 10,
					},
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On("SearchWithKey", ctx, db, domain.ExternalConfigSearchArgs{
					Keyword: "not-found-key-2", Limit: int64(10), Offset: int64(0),
				}).
					Return([]*domain.ExternalConfiguration{}, nil).Once()
			},
		},
		{
			name: "internal err",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetExternalConfigurationsRequest{Keyword: "some-key-2", Paging: &cpb.Paging{
				Limit: 10,
			}},
			expectedErr:  status.Error(codes.Internal, "internal err"),
			expectedResp: &mpb.GetExternalConfigurationsResponse{},
			setup: func(ctx context.Context) {
				configRepo.On("SearchWithKey", ctx, db, domain.ExternalConfigSearchArgs{
					Keyword: "some-key-2", Limit: int64(10), Offset: int64(0),
				}).
					Return([]*domain.ExternalConfiguration{}, errors.New("internal err")).Once()
			},
		},
	}

	for _, testCase := range tc {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetExternalConfigurationsRequest)
			resp, err := s.GetExternalConfigurations(testCase.ctx, req)
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

func TestConfigurationService_GetExternalConfigurationByKey(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	configRepo := &mock_repo.MockExternalConfigRepo{}
	existing := make([]*domain.ExternalConfiguration, 10)
	for i := 0; i < 10; i++ {
		id := idutil.ULIDNow()
		existing[i] = &domain.ExternalConfiguration{
			ID:          id,
			ConfigKey:   "z-key-" + strconv.Itoa(i) + id,
			ConfigValue: "value-" + strconv.Itoa(i) + id,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	s := NewExternalConfigurationService(db, configRepo)

	tc := []TestCase{
		{
			name:        "config found",
			ctx:         interceptors.ContextWithUserID(ctx, "id"),
			req:         &mpb.GetExternalConfigurationByKeyRequest{Key: existing[2].ConfigKey},
			expectedErr: nil,
			expectedResp: &mpb.GetExternalConfigurationByKeyResponse{
				Configuration: &mpb.ExternalConfiguration{
					Id:          existing[2].ID,
					ConfigKey:   existing[2].ConfigKey,
					ConfigValue: existing[2].ConfigValue,
					CreatedAt:   existing[2].CreatedAt.String(),
					UpdatedAt:   existing[2].UpdatedAt.String(),
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, db, existing[2].ConfigKey).
					Return(&domain.ExternalConfiguration{
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
			req:          &mpb.GetExternalConfigurationByKeyRequest{Key: "not-found-key-2"},
			expectedErr:  status.Error(codes.NotFound, pgx.ErrNoRows.Error()),
			expectedResp: &mpb.GetExternalConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, db, "not-found-key-2").
					Return(nil, pgx.ErrNoRows).Times(1)
			},
		},
		{
			name:         "empty key",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.GetExternalConfigurationByKeyRequest{Key: "  "},
			expectedErr:  status.Error(codes.FailedPrecondition, "configuration key cannot be empty"),
			expectedResp: &mpb.GetExternalConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:         "internal err",
			ctx:          interceptors.ContextWithUserID(ctx, "id"),
			req:          &mpb.GetExternalConfigurationByKeyRequest{Key: "some-key"},
			expectedErr:  status.Error(codes.Internal, "internal err"),
			expectedResp: &mpb.GetExternalConfigurationByKeyResponse{},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKey", ctx, db, "some-key").
					Return(nil, errors.New("internal err")).Times(1)
			},
		},
	}

	for _, testCase := range tc {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetExternalConfigurationByKeyRequest)
			resp, err := s.GetExternalConfigurationByKey(testCase.ctx, req)
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

func TestConfigurationService_CreateMultiConfigurations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	configRepo := &mock_repo.MockExternalConfigRepo{}

	s := NewExternalConfigurationService(db, configRepo)
	externalConfigs := []*mpb.CreateMultiConfigurationsRequest_ExternalConfiguration{
		{Key: "Key", Value: "Value", ValueType: "ValueType"},
	}
	tc := []TestCase{
		{
			name: "create success",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.CreateMultiConfigurationsRequest{
				ExternalConfigurations: externalConfigs,
			},
			expectedErr: nil,
			expectedResp: &mpb.CreateMultiConfigurationsResponse{
				Successful: true,
			},
			setup: func(ctx context.Context) {

				configRepo.On("CreateMultipleConfigs", ctx, db, mock.Anything).
					Return(nil).Times(1)
			},
		},
		{
			name: "internal err",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.CreateMultiConfigurationsRequest{
				ExternalConfigurations: externalConfigs,
			},
			expectedErr: status.Error(codes.Internal, "internal err"),
			expectedResp: &mpb.CreateMultiConfigurationsResponse{
				Successful: false,
			},
			setup: func(ctx context.Context) {
				configRepo.On("CreateMultipleConfigs", ctx, db, mock.Anything).
					Return(errors.New("internal err")).Times(1)
			},
		},
	}

	for _, testCase := range tc {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.CreateMultiConfigurationsRequest)
			resp, err := s.CreateMultiConfigurations(testCase.ctx, req)
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

func TestConfigurationService_GetConfigurationByKeysAndLocations(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	configRepo := &mock_repo.MockExternalConfigRepo{}
	existing := make([]*domain.LocationConfiguration, 10)
	for i := 0; i < 5; i++ {
		id := strconv.Itoa(i)
		existing[i] = &domain.LocationConfiguration{
			ID:          id,
			ConfigKey:   "z-key-" + id,
			LocationID:  "location-" + id,
			ConfigValue: "value-" + id,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	s := NewExternalConfigurationService(db, configRepo)

	tc := []TestCase{
		{
			name: "find config by 1 key and 1 location",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsRequest{
				Keys:         []string{existing[2].ConfigKey},
				LocationsIds: []string{existing[2].LocationID},
			},
			expectedErr: nil,
			expectedResp: &mpb.GetConfigurationByKeysAndLocationsResponse{
				Configurations: []*mpb.LocationConfiguration{
					{
						Id:          existing[2].ID,
						ConfigKey:   existing[2].ConfigKey,
						LocationId:  existing[2].LocationID,
						ConfigValue: existing[2].ConfigValue,
						CreatedAt:   timestamppb.New(existing[2].CreatedAt),
						UpdatedAt:   timestamppb.New(existing[2].UpdatedAt),
					},
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKeysAndLocations", ctx, db, []string{existing[2].ConfigKey}, []string{existing[2].LocationID}).
					Return([]*domain.LocationConfiguration{existing[2]}, nil).Times(1)
			},
		},
		{
			name: "find config by many keys and many locations",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsRequest{
				Keys:         []string{existing[0].ConfigKey, existing[1].ConfigKey},
				LocationsIds: []string{existing[0].LocationID, existing[1].LocationID},
			},
			expectedErr: nil,
			expectedResp: &mpb.GetConfigurationByKeysAndLocationsResponse{
				Configurations: []*mpb.LocationConfiguration{
					{
						Id:          existing[0].ID,
						ConfigKey:   existing[0].ConfigKey,
						LocationId:  existing[0].LocationID,
						ConfigValue: existing[0].ConfigValue,
						CreatedAt:   timestamppb.New(existing[0].CreatedAt),
						UpdatedAt:   timestamppb.New(existing[0].UpdatedAt),
					},
					{
						Id:          existing[1].ID,
						ConfigKey:   existing[1].ConfigKey,
						LocationId:  existing[1].LocationID,
						ConfigValue: existing[1].ConfigValue,
						CreatedAt:   timestamppb.New(existing[1].CreatedAt),
						UpdatedAt:   timestamppb.New(existing[1].UpdatedAt),
					},
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On(
					"GetByKeysAndLocations",
					ctx,
					db,
					[]string{existing[0].ConfigKey, existing[1].ConfigKey},
					[]string{existing[0].LocationID, existing[1].LocationID},
				).
					Return([]*domain.LocationConfiguration{existing[0], existing[1]}, nil).Times(1)
			},
		},
		{
			name: "miss keys field",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsRequest{
				LocationsIds: []string{existing[2].LocationID},
			},
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("keys and location ids are required fields in request")),
			expectedResp: nil,
			setup:        func(ctx context.Context) {},
		},
		{
			name: "miss location ids field",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsRequest{
				Keys: []string{existing[2].ConfigKey},
			},
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("keys and location ids are required fields in request")),
			expectedResp: nil,
			setup:        func(ctx context.Context) {},
		},
		{
			name: "internal error",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsRequest{
				Keys:         []string{existing[2].ConfigKey},
				LocationsIds: []string{existing[2].LocationID},
			},
			expectedErr:  status.Error(codes.Internal, "ConfigRepo.GetByKeysAndLocations: internal err"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				configRepo.On("GetByKeysAndLocations", ctx, db, []string{existing[2].ConfigKey}, []string{existing[2].LocationID}).
					Return(nil, errors.New("internal err")).Times(1)
			},
		},
	}

	for _, testCase := range tc {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetConfigurationByKeysAndLocationsRequest)
			resp, err := s.GetConfigurationByKeysAndLocations(testCase.ctx, req)
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

func TestConfigurationService_GetConfigurationByKeysAndLocationsV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	now := time.Now().UTC()
	configRepo := &mock_repo.MockExternalConfigRepo{}
	existing := make([]*domain.LocationConfigurationV2, 10)
	for i := 0; i < 5; i++ {
		id := strconv.Itoa(i)
		existing[i] = &domain.LocationConfigurationV2{
			ID:          id,
			ConfigKey:   "z-key-" + id,
			LocationID:  "location-" + id,
			ConfigValue: "value-" + id,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}
	s := NewExternalConfigurationService(db, configRepo)

	tc := []TestCase{
		{
			name: "find config by 1 key and 1 location",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsV2Request{
				Keys:        []string{existing[2].ConfigKey},
				LocationIds: []string{existing[2].LocationID},
			},
			expectedErr: nil,
			expectedResp: &mpb.GetConfigurationByKeysAndLocationsV2Response{
				Configurations: []*mpb.LocationConfiguration{
					{
						Id:          existing[2].ID,
						ConfigKey:   existing[2].ConfigKey,
						LocationId:  existing[2].LocationID,
						ConfigValue: existing[2].ConfigValue,
						CreatedAt:   timestamppb.New(existing[2].CreatedAt),
						UpdatedAt:   timestamppb.New(existing[2].UpdatedAt),
					},
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On("GetByKeysAndLocationsV2", ctx, db, []string{existing[2].ConfigKey}, []string{existing[2].LocationID}).
					Return([]*domain.LocationConfigurationV2{existing[2]}, nil).Times(1)
			},
		},
		{
			name: "find config by many keys and many locations",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsV2Request{
				Keys:        []string{existing[0].ConfigKey, existing[1].ConfigKey},
				LocationIds: []string{existing[0].LocationID, existing[1].LocationID},
			},
			expectedErr: nil,
			expectedResp: &mpb.GetConfigurationByKeysAndLocationsV2Response{
				Configurations: []*mpb.LocationConfiguration{
					{
						Id:          existing[0].ID,
						ConfigKey:   existing[0].ConfigKey,
						LocationId:  existing[0].LocationID,
						ConfigValue: existing[0].ConfigValue,
						CreatedAt:   timestamppb.New(existing[0].CreatedAt),
						UpdatedAt:   timestamppb.New(existing[0].UpdatedAt),
					},
					{
						Id:          existing[1].ID,
						ConfigKey:   existing[1].ConfigKey,
						LocationId:  existing[1].LocationID,
						ConfigValue: existing[1].ConfigValue,
						CreatedAt:   timestamppb.New(existing[1].CreatedAt),
						UpdatedAt:   timestamppb.New(existing[1].UpdatedAt),
					},
				},
			},
			setup: func(ctx context.Context) {
				configRepo.On(
					"GetByKeysAndLocationsV2",
					ctx,
					db,
					[]string{existing[0].ConfigKey, existing[1].ConfigKey},
					[]string{existing[0].LocationID, existing[1].LocationID},
				).
					Return([]*domain.LocationConfigurationV2{existing[0], existing[1]}, nil).Times(1)
			},
		},
		{
			name: "miss keys field",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsV2Request{
				LocationIds: []string{existing[2].LocationID},
			},
			expectedErr:  status.Error(codes.InvalidArgument, fmt.Sprintf("configuration key is required")),
			expectedResp: nil,
			setup:        func(ctx context.Context) {},
		},
		{
			name: "internal error",
			ctx:  interceptors.ContextWithUserID(ctx, "id"),
			req: &mpb.GetConfigurationByKeysAndLocationsV2Request{
				Keys:        []string{existing[2].ConfigKey},
				LocationIds: []string{existing[2].LocationID},
			},
			expectedErr:  status.Error(codes.Internal, "ConfigRepo.GetByKeysAndLocationsV2: internal err"),
			expectedResp: nil,
			setup: func(ctx context.Context) {
				configRepo.On("GetByKeysAndLocationsV2", ctx, db, []string{existing[2].ConfigKey}, []string{existing[2].LocationID}).
					Return(nil, errors.New("internal err")).Times(1)
			},
		},
	}

	for _, testCase := range tc {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*mpb.GetConfigurationByKeysAndLocationsV2Request)
			resp, err := s.GetConfigurationByKeysAndLocationsV2(testCase.ctx, req)
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
