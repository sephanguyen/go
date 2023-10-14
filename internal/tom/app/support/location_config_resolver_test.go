package support

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	tom_const "github.com/manabie-com/backend/internal/tom/constants"
	"github.com/manabie-com/backend/mock/testutil"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type ExternalConfigurationServiceMock struct {
	getConfigurationByKeysAndLocations   func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error)
	getConfigurationByKeysAndLocationsV2 func(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error)
}

func (e *ExternalConfigurationServiceMock) GetConfigurationByKeysAndLocations(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error) {
	return e.getConfigurationByKeysAndLocations(ctx, in, opts...)
}

func (e *ExternalConfigurationServiceMock) GetConfigurationByKeysAndLocationsV2(ctx context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
	return e.getConfigurationByKeysAndLocationsV2(ctx, in, opts...)
}

func Test_GetEnabledLocationConfigsByLocations(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	locationRepo := &mock_repositories.MockLocationRepo{}
	externalCfgService := &ExternalConfigurationServiceMock{}
	s := &LocationConfigResolver{
		DB:                           mockDB.DB,
		LocationRepo:                 locationRepo,
		ExternalConfigurationService: externalCfgService,
	}

	studentConfigKey := tom_const.ChatConfigKeyStudentV2
	parentConfigKey := tom_const.ChatConfigKeyParentV2

	testCases := []struct {
		Name                     string
		Ctx                      context.Context
		LocationIDs              []string
		RequestConversationTypes []tpb.ConversationType
		Err                      error
		ExpectRes                map[tpb.ConversationType][]string
		Setup                    func(ctx context.Context)
	}{
		{
			Name: "admin requests",
			Ctx: func() context.Context {
				claims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						SchoolIDs: []string{"1"},
					},
				}
				ctx := context.Background()
				ctx = interceptors.ContextWithUserID(ctx, "user")
				ctx = interceptors.ContextWithJWTClaims(ctx, claims)
				return ctx
			}(),
			LocationIDs: []string{},
			RequestConversationTypes: []tpb.ConversationType{
				tpb.ConversationType_CONVERSATION_STUDENT,
				tpb.ConversationType_CONVERSATION_PARENT,
			},
			Err: nil,
			ExpectRes: map[tpb.ConversationType][]string{
				tpb.ConversationType_CONVERSATION_STUDENT: {
					"loc/lowest-loc-1",
				},
				tpb.ConversationType_CONVERSATION_PARENT: {
					"loc/lowest-loc-1",
					"loc/lowest-loc-2",
				},
			},
			Setup: func(_ context.Context) {
				rootIDs := []string{"root-id-1", "root-id-2"}
				locationRepo.On("FindRootIDs", mock.Anything, mock.Anything).Once().Return(
					rootIDs,
					nil,
				)
				lowestLocationIDs := []string{"lowest-loc-1", "lowest-loc-2"}
				lowestAccessPath := map[string]string{
					lowestLocationIDs[0]: "loc/lowest-loc-1",
					lowestLocationIDs[1]: "loc/lowest-loc-2",
				}
				locationRepo.On("FindLowestAccessPathByLocationIDs", mock.Anything, mock.Anything, rootIDs).Once().Return(
					lowestLocationIDs,
					lowestAccessPath,
					nil,
				)

				// return this config
				// studentConfig location 1 = true
				// studentConfig location 2 = false
				// parentConfig location 1 = true
				// parentConfig location 2 = true
				externalCfgService.getConfigurationByKeysAndLocationsV2 = func(_ context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{studentConfigKey, parentConfigKey})
					assert.ElementsMatch(t, in.LocationIds, lowestLocationIDs)
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      lowestLocationIDs[0],
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      lowestLocationIDs[1],
								ConfigValue:     "false",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       parentConfigKey,
								LocationId:      lowestLocationIDs[0],
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       parentConfigKey,
								LocationId:      lowestLocationIDs[1],
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
			},
		},
		{
			Name: "staff requests",
			Ctx: func() context.Context {
				claims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						SchoolIDs: []string{"1"},
					},
				}
				ctx := context.Background()
				ctx = interceptors.ContextWithUserID(ctx, "user")
				ctx = interceptors.ContextWithJWTClaims(ctx, claims)
				return ctx
			}(),
			LocationIDs: []string{"lowest-loc-1", "lowest-loc-2"},
			RequestConversationTypes: []tpb.ConversationType{
				tpb.ConversationType_CONVERSATION_STUDENT,
				tpb.ConversationType_CONVERSATION_PARENT,
			},
			Err: nil,
			ExpectRes: map[tpb.ConversationType][]string{
				tpb.ConversationType_CONVERSATION_STUDENT: {
					"loc/lowest-loc-1",
				},
				tpb.ConversationType_CONVERSATION_PARENT: {
					"loc/lowest-loc-1",
					"loc/lowest-loc-2",
				},
			},
			Setup: func(_ context.Context) {
				rootIDs := []string{"lowest-loc-1", "lowest-loc-2"}
				lowestLocationIDs := []string{"lowest-loc-1", "lowest-loc-2"}
				lowestAccessPath := map[string]string{
					lowestLocationIDs[0]: "loc/lowest-loc-1",
					lowestLocationIDs[1]: "loc/lowest-loc-2",
				}
				locationRepo.On("FindLowestAccessPathByLocationIDs", mock.Anything, mock.Anything, rootIDs).Once().Return(
					lowestLocationIDs,
					lowestAccessPath,
					nil,
				)

				// return this config
				// studentConfig location 1 = true
				// studentConfig location 2 = false
				// parentConfig location 1 = true
				// parentConfig location 2 = true
				externalCfgService.getConfigurationByKeysAndLocationsV2 = func(_ context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{studentConfigKey, parentConfigKey})
					assert.ElementsMatch(t, in.LocationIds, lowestLocationIDs)
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      lowestLocationIDs[0],
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      lowestLocationIDs[1],
								ConfigValue:     "false",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       parentConfigKey,
								LocationId:      lowestLocationIDs[0],
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       parentConfigKey,
								LocationId:      lowestLocationIDs[1],
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
			},
		},
		{
			Name: "staff requests only student",
			Ctx: func() context.Context {
				claims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						SchoolIDs: []string{"1"},
					},
				}
				ctx := context.Background()
				ctx = interceptors.ContextWithUserID(ctx, "user")
				ctx = interceptors.ContextWithJWTClaims(ctx, claims)
				return ctx
			}(),
			LocationIDs: []string{"lowest-loc-1", "lowest-loc-2"},
			RequestConversationTypes: []tpb.ConversationType{
				tpb.ConversationType_CONVERSATION_STUDENT,
			},
			Err: nil,
			ExpectRes: map[tpb.ConversationType][]string{
				tpb.ConversationType_CONVERSATION_STUDENT: {
					"loc/lowest-loc-1",
				},
			},
			Setup: func(_ context.Context) {
				rootIDs := []string{"lowest-loc-1", "lowest-loc-2"}
				lowestLocationIDs := []string{"lowest-loc-1", "lowest-loc-2"}
				lowestAccessPath := map[string]string{
					lowestLocationIDs[0]: "loc/lowest-loc-1",
					lowestLocationIDs[1]: "loc/lowest-loc-2",
				}
				locationRepo.On("FindLowestAccessPathByLocationIDs", mock.Anything, mock.Anything, rootIDs).Once().Return(
					lowestLocationIDs,
					lowestAccessPath,
					nil,
				)

				// return this config
				// studentConfig location 1 = true
				// studentConfig location 2 = false
				externalCfgService.getConfigurationByKeysAndLocationsV2 = func(_ context.Context, in *mpb.GetConfigurationByKeysAndLocationsV2Request, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsV2Response, error) {
					assert.ElementsMatch(t, in.Keys, []string{studentConfigKey})
					assert.ElementsMatch(t, in.LocationIds, lowestLocationIDs)
					return &mpb.GetConfigurationByKeysAndLocationsV2Response{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      lowestLocationIDs[0],
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      lowestLocationIDs[1],
								ConfigValue:     "false",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
			},
		},
	}

	for _, tc := range testCases {
		tc.Setup(tc.Ctx)
		tc.Ctx = interceptors.NewIncomingContext(tc.Ctx)
		res, err := s.GetEnabledLocationConfigsByLocations(tc.Ctx, tc.LocationIDs, tc.RequestConversationTypes)
		assert.Equal(t, tc.Err, err)
		assert.Equal(t, tc.ExpectRes, res)
	}
}

func Test_GetEnabledLocationConfigsByOrg(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	locationRepo := &mock_repositories.MockLocationRepo{}
	externalCfgService := &ExternalConfigurationServiceMock{}
	s := &LocationConfigResolver{
		DB:                           mockDB.DB,
		LocationRepo:                 locationRepo,
		ExternalConfigurationService: externalCfgService,
	}

	studentConfigKey := tom_const.ChatConfigKeyStudent
	parentConfigKey := tom_const.ChatConfigKeyParent

	testCases := []struct {
		Name              string
		Ctx               context.Context
		LocationIDs       []string
		Err               error
		ExpectExcludeType []tpb.ConversationType
		ExpectAccessPaths []string
		Setup             func(ctx context.Context)
	}{
		{
			Name: "admin requests",
			Ctx: func() context.Context {
				claims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						SchoolIDs: []string{"1"},
					},
				}
				ctx := context.Background()
				ctx = interceptors.ContextWithUserID(ctx, "user")
				ctx = interceptors.ContextWithJWTClaims(ctx, claims)
				return ctx
			}(),
			LocationIDs:       []string{},
			Err:               nil,
			ExpectExcludeType: []tpb.ConversationType{},
			ExpectAccessPaths: nil,
			Setup: func(_ context.Context) {
				rootIDs := []string{"root-id-1", "root-id-2"}
				locationRepo.On("FindRootIDs", mock.Anything, mock.Anything).Once().Return(
					rootIDs,
					nil,
				)

				// return this config
				// studentConfig location 1 = true
				// studentConfig location 2 = false
				// parentConfig location 1 = true
				// parentConfig location 2 = true
				externalCfgService.getConfigurationByKeysAndLocations = func(_ context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error) {
					assert.ElementsMatch(t, in.Keys, []string{studentConfigKey, parentConfigKey})
					assert.ElementsMatch(t, in.LocationsIds, []string{rootIDs[0]})
					return &mpb.GetConfigurationByKeysAndLocationsResponse{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      "",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       parentConfigKey,
								LocationId:      "",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
			},
		},
		{
			Name: "staff requests",
			Ctx: func() context.Context {
				claims := &interceptors.CustomClaims{
					Manabie: &interceptors.ManabieClaims{
						SchoolIDs: []string{"1"},
					},
				}
				ctx := context.Background()
				ctx = interceptors.ContextWithUserID(ctx, "user")
				ctx = interceptors.ContextWithJWTClaims(ctx, claims)
				return ctx
			}(),
			LocationIDs: []string{
				"brand",
				"centre1",
				"centre2",
			},
			Err: nil,
			ExpectExcludeType: []tpb.ConversationType{
				tpb.ConversationType_CONVERSATION_PARENT,
			},
			ExpectAccessPaths: []string{
				"org/brand",
				"org/brand/centre1",
				"org/brand/centre2",
			},
			Setup: func(_ context.Context) {
				locationIDs := []string{"brand", "centre1", "centre2"}
				rootLocation := "org"
				accessPathInDB := []string{
					rootLocation + "/" + locationIDs[0],
					rootLocation + "/" + locationIDs[0] + "/" + locationIDs[1],
					rootLocation + "/" + locationIDs[0] + "/" + locationIDs[2],
				}
				locationRepo.On("FindAccessPaths", mock.Anything, mock.Anything, locationIDs).Once().Return(
					accessPathInDB,
					nil,
				)

				// return this config
				// studentConfig location 1 = true
				// parentConfig location 1 = true
				externalCfgService.getConfigurationByKeysAndLocations = func(_ context.Context, in *mpb.GetConfigurationByKeysAndLocationsRequest, opts ...grpc.CallOption) (*mpb.GetConfigurationByKeysAndLocationsResponse, error) {
					assert.ElementsMatch(t, in.Keys, []string{studentConfigKey, parentConfigKey})
					assert.ElementsMatch(t, in.LocationsIds, []string{rootLocation})
					return &mpb.GetConfigurationByKeysAndLocationsResponse{
						Configurations: []*mpb.LocationConfiguration{
							{
								Id:              "id",
								ConfigKey:       studentConfigKey,
								LocationId:      "org",
								ConfigValue:     "true",
								ConfigValueType: "boolean",
							},
							{
								Id:              "id",
								ConfigKey:       parentConfigKey,
								LocationId:      "org",
								ConfigValue:     "false",
								ConfigValueType: "boolean",
							},
						},
					}, nil
				}
			},
		},
	}

	for _, tc := range testCases {
		tc.Setup(tc.Ctx)
		tc.Ctx = interceptors.NewIncomingContext(tc.Ctx)
		excludeConvType, accessPaths, err := s.GetEnabledLocationConfigsByOrg(tc.Ctx, tc.LocationIDs)
		assert.Equal(t, tc.Err, err)
		assert.Equal(t, tc.ExpectExcludeType, excludeConvType)
		assert.Equal(t, tc.ExpectAccessPaths, accessPaths)
	}
}
