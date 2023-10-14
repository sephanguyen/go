package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TestCase struct {
	name             string
	ctx              context.Context
	req              interface{}
	expectedResp     interface{}
	expectedErr      error
	setup            func(ctx context.Context)
	expectedErrModel *errdetails.BadRequest
}

func TestLocationManagementGRPCService_ImportLocationTypeV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	locationRepo := new(mock_repositories.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	locationTypeRepo := new(mock_repositories.MockLocationTypeRepo)
	importLogRepo := new(mock_repositories.MockImportLogRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := NewLocationManagementGRPCService(
		db,
		jsm,
		locationRepo,
		locationTypeRepo,
		importLogRepo,
		mockUnleashClient,
		"stag",
	)

	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportLocationTypeV2Request{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - number of column != 3",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 3, got 2"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name
							1,LocType 1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be name, got namez"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`namez,display_name,level
							1,LocType 1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != display_name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be display_name, got display_namez"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_namez,level
							1,LocType 1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != level",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be level, got levelz"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,levelz
							1,LocType 1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file with invalid values",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(fmt.Sprintf(`name,display_name,level
							org1,LocType 1,1,bool
							org2,   ,2
							org,Loc type xyz,3
							%s`, fmt.Sprintf("%s,display,4,1", string([]byte{0xff, 0xfe, 0xfd})))),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "bool is not a valid boolean: strconv.ParseBool: parsing \"bool\": invalid syntax",
					},
					{
						Field:       "Row Number: 3",
						Description: "display name can not be empty",
					},
					{
						Field:       "Row Number: 4",
						Description: "can not import org",
					},
					{
						Field:       "Row Number: 5",
						Description: `name is not a valid UTF8 string`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file with duplication name and level",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
						org1,LocType 1,1,1
						org2,Loc Type 2,2
						org3,Loc type 3,3
						org4,Loc type 4,3`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 3",
						Description: "name org is duplicated",
					},
					{
						Field:       "Row Number: 5",
						Description: `level 3 is duplicated`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file with wrong level order",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
					org1,LocType 1,-1,1
					org2,Loc Type 2,4
					org3,Loc type 3,3`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "level must be greater than 0",
					},
					{
						Field:       "Row Number: 4",
						Description: `level must be in sequential order`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportLocationTypeV2(testCase.ctx, testCase.req.(*mpb.ImportLocationTypeV2Request))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, testCase.expectedErr, err)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, locationTypeRepo)
			}
		})
	}
}

func TestLocationManagementGRPCService_ImportLocationTypeV2_NewRule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	locationRepo := new(mock_repositories.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	locationTypeRepo := new(mock_repositories.MockLocationTypeRepo)
	importLogRepo := new(mock_repositories.MockImportLogRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := NewLocationManagementGRPCService(
		db,
		jsm,
		locationRepo,
		locationTypeRepo,
		importLogRepo,
		mockUnleashClient,
		"stag",
	)

	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportLocationTypeV2Request{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - number of column != 3",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 3, got 2"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name
							1,LocType 1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be name, got namez"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`namez,display_name,level
							1,LocType 1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != display_name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be display_name, got display_namez"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_namez,level
							1,LocType 1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != level",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be level, got levelz"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,levelz
							1,LocType 1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file with invalid values",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(fmt.Sprintf(`name,display_name,level
							org1,LocType 1,1,bool
							org2,   ,2
							org,Loc type xyz,3
							%s`, fmt.Sprintf("%s,display,4,1", string([]byte{0xff, 0xfe, 0xfd})))),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "bool is not a valid boolean: strconv.ParseBool: parsing \"bool\": invalid syntax",
					},
					{
						Field:       "Row Number: 3",
						Description: "display name can not be empty",
					},
					{
						Field:       "Row Number: 4",
						Description: "can not import org",
					},
					{
						Field:       "Row Number: 5",
						Description: `name is not a valid UTF8 string`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file with duplication name and level",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
						org1,LocType 1,1,1
						org2,Loc Type 2,2
						org3,Loc type 3,3
						org4,Loc type 4,3`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 3",
						Description: "name org is duplicated",
					},
					{
						Field:       "Row Number: 5",
						Description: `level 3 is duplicated`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file with wrong level order",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
					org1,LocType 1,-1,1
					org2,Loc Type 2,4
					org3,Loc type 3,3`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "level must be greater than 0",
					},
					{
						Field:       "Row Number: 4",
						Description: `level must be in sequential order`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "parsing valid file with not sequential order",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
					org1,LocType 1,1
					org2,Loc Type 2,2
					org3,Loc type 3,4`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 4",
						Description: `level must be in sequential order`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "valid file but violate rule mustImportAllExistData",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
					brand,brand,1`),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("resources.masters.message.mustImportAllExistData").Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				locationTypeRepo.On("GetAllLocationTypes", ctx, db, mock.Anything).
					Return([]*repo.LocationType{
						{
							LocationTypeID: database.Text("location-type-id"),
							Name:           database.Text("org"),
							Level:          database.Int4(0),
						},
						{
							LocationTypeID: database.Text("location-type-id2"),
							Name:           database.Text("brand"),
							Level:          database.Int4(1),
						},
						{
							LocationTypeID: database.Text("location-type-id22"),
							Name:           database.Text("center"),
							Level:          database.Int4(2),
						},
					}, nil).
					Once()
				locationTypeRepo.On("Import", ctx, db, mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "valid file but violate rule canNotUpdateLowestType",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
					brand,brand,1
					center,center,2
					center1,center1,3`),
			},
			expectedErr: status.Error(codes.InvalidArgument, fmt.Errorf("resources.masters.message.canNotUpdateLowestType").Error()),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				locationTypeRepo.On("GetAllLocationTypes", ctx, db, mock.Anything).
					Return([]*repo.LocationType{
						{
							LocationTypeID: database.Text("location-type-id"),
							Name:           database.Text("org"),
							Level:          database.Int4(0),
						},
						{
							LocationTypeID: database.Text("location-type-id2"),
							Name:           database.Text("brand"),
							Level:          database.Int4(1),
						},
						{
							LocationTypeID: database.Text("location-type-id22"),
							Name:           database.Text("center"),
							Level:          database.Int4(2),
						},
					}, nil).
					Once()
				locationTypeRepo.On("Import", ctx, db, mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "valid file with valid values should be imported",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
					brand,brand,1
					center,center,2`),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				locationTypeRepo.On("GetAllLocationTypes", ctx, db, mock.Anything).
					Return([]*repo.LocationType{
						{
							LocationTypeID: database.Text("location-type-id"),
							Name:           database.Text("org"),
							Level:          database.Int4(0),
						},
						{
							LocationTypeID: database.Text("location-type-id2"),
							Name:           database.Text("brand"),
							Level:          database.Int4(1),
						},
						{
							LocationTypeID: database.Text("location-type-id22"),
							Name:           database.Text("center"),
							Level:          database.Int4(2),
						},
					}, nil).
					Once()
				locationTypeRepo.On("Import", ctx, db, mock.Anything).Return(nil)
				tx.On("Commit", ctx).Once().Return(nil)
			},
		},
		{
			name: "valid file with middle value should be imported",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationTypeV2Request{
				Payload: []byte(`name,display_name,level
					brand,brand,1
					center1,center1,2
					center,center,3`),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				locationTypeRepo.On("GetAllLocationTypes", ctx, db, mock.Anything).
					Return([]*repo.LocationType{
						{
							LocationTypeID: database.Text("location-type-id"),
							Name:           database.Text("org"),
							Level:          database.Int4(0),
						},
						{
							LocationTypeID: database.Text("location-type-id2"),
							Name:           database.Text("brand"),
							Level:          database.Int4(1),
						},
						{
							LocationTypeID: database.Text("location-type-id22"),
							Name:           database.Text("center"),
							Level:          database.Int4(2),
						},
					}, nil).
					Once()
				locationTypeRepo.On("Import", ctx, db, mock.Anything).Return(nil)
				tx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportLocationTypeV2(testCase.ctx, testCase.req.(*mpb.ImportLocationTypeV2Request))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, testCase.expectedErr, err)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, locationTypeRepo)
			}
		})
	}
}

func TestLocationManagementGRPCService_ImportLocationV2(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	locationRepo := new(mock_repositories.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	locationTypeRepo := new(mock_repositories.MockLocationTypeRepo)
	importLogRepo := new(mock_repositories.MockImportLogRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := NewLocationManagementGRPCService(
		db,
		jsm,
		locationRepo,
		locationTypeRepo,
		importLogRepo,
		mockUnleashClient,
		"",
	)

	testCases := []TestCase{
		{
			name:        "no data in csv file",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "no data in csv file"),
			req:         &mpb.ImportLocationV2Request{},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - number of column != 4",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "wrong number of columns, expected 4, got 2"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`name,location_type
							1,LocType 1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - first column name (toLowerCase) != partner_internal_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 1 should be partner_internal_id, got partner_internal_idz"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_idz,name,location_type,partner_internal_parent_id
							name,LocType 1,1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - second column name (toLowerCase) != name",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 2 should be name, got namez"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,namez,location_type,partner_internal_parent_id
							name,LocType 1,1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - third column name (toLowerCase) != location_type",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 3 should be location_type, got location_typez"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,name,location_typez,partner_internal_parent_id
							name,LocType 1,1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "invalid file - fourth column name (toLowerCase) != partner_internal_parent_id",
			ctx:         interceptors.ContextWithUserID(ctx, "user-id"),
			expectedErr: status.Error(codes.InvalidArgument, "csv has invalid format, column number 4 should be partner_internal_parent_id, got partner_internal_parent_idz"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,name,location_type,partner_internal_parent_idz
							name,LocType 1,1,1`),
			},
			setup: func(ctx context.Context) {
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportLocationV2(testCase.ctx, testCase.req.(*mpb.ImportLocationV2Request))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, testCase.expectedErr, err)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, locationTypeRepo)
			}
		})
	}
}

func TestLocationManagementGRPCService_ImportLocationV2_NewRule(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db := new(mock_database.Ext)
	tx := &mock_database.Tx{}
	locationRepo := new(mock_repositories.MockLocationRepo)
	jsm := new(mock_nats.JetStreamManagement)
	locationTypeRepo := new(mock_repositories.MockLocationTypeRepo)
	importLogRepo := new(mock_repositories.MockImportLogRepo)

	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := NewLocationManagementGRPCService(
		db,
		jsm,
		locationRepo,
		locationTypeRepo,
		importLogRepo,
		mockUnleashClient,
		"",
	)

	testCases := []TestCase{
		{
			name: "parsing valid file with invalid values",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(fmt.Sprintf(`partner_internal_id,name,location_type,partner_internal_parent_id
				pID1,Location 1,location-type-1,location2,yes
				pIDA,,center,12
				pIDA,Location 16,locType,
				pIDB,Location 1,brand,
				%s`, fmt.Sprintf("pIDC,%s,brand,", string([]byte{0xff, 0xfe, 0xfd})))),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 2",
						Description: "yes is not a valid boolean: strconv.ParseBool: parsing \"yes\": invalid syntax",
					},
					{
						Field:       "Row Number: 3",
						Description: "column name is required",
					},
					{
						Field:       "Row Number: 4",
						Description: "partner internal id pIDA is duplicated",
					},
					{
						Field:       "Row Number: 6",
						Description: `name is not a valid UTF8 string`,
					},
				},
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "valid file but violate policy mustImportAllExistData",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,name,location_type,partner_internal_parent_id
				loc_brand,brand,brand,`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "resources.masters.message.mustImportAllExistData"),
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				locationRepo.On("GetLocationByLocationTypeName", ctx, db, "org").
					Return([]*domain.Location{
						{
							LocationID:        "location-id-0",
							LocationType:      "org",
							PartnerInternalID: "partner A",
						},
					}, nil).
					Once()
				locationTypeRepo.On("RetrieveLocationTypes", ctx, db).
					Once().
					Return([]*domain.LocationType{
						{
							LocationTypeID: "org",
							Name:           "org",
							Level:          0,
						},
						{
							LocationTypeID: "brand",
							Name:           "brand",
							Level:          1,
						},
						{
							LocationTypeID: "center",
							Name:           "center",
							Level:          2,
						},
						{
							LocationTypeID: "area",
							Name:           "area",
							Level:          3,
						},
					}, nil)
				locationRepo.On("GetAllRawLocations", ctx, db).
					Once().
					Return([]*domain.Location{
						{
							LocationID:        "location-org",
							PartnerInternalID: "loc_org",
							LocationType:      "org",
						},
						{
							LocationID:              "location-brand",
							PartnerInternalID:       "loc_brand",
							LocationType:            "brand",
							PartnerInternalParentID: "",
						},
						{
							LocationID:              "location-center",
							PartnerInternalID:       "loc_center",
							LocationType:            "center",
							PartnerInternalParentID: "loc_brand",
						},
					}, nil)

				tx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "valid file but violate policy all location should not no child",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,name,location_type,partner_internal_parent_id
				brand,brand,brand,
				loc_center,center,center,brand
				brand1,brand1,brand,`),
			},
			expectedErr: status.Error(codes.InvalidArgument, "data is not valid, please check"),
			expectedErrModel: &errdetails.BadRequest{
				FieldViolations: []*errdetails.BadRequest_FieldViolation{
					{
						Field:       "Row Number: 4",
						Description: "cannot import location which is parent having no child",
					},
				},
			},
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				locationRepo.On("GetLocationByLocationTypeName", ctx, db, "org").
					Return([]*domain.Location{
						{
							LocationID:        "location-id-0",
							LocationType:      "org",
							PartnerInternalID: "partner A",
						},
					}, nil).
					Once()
				locationTypeRepo.On("RetrieveLocationTypes", ctx, db).
					Once().
					Return([]*domain.LocationType{
						{
							LocationTypeID: "org",
							Name:           "org",
							Level:          0,
						},
						{
							LocationTypeID: "brand",
							Name:           "brand",
							Level:          1,
						},
						{
							LocationTypeID: "center",
							Name:           "center",
							Level:          2,
						},
					}, nil)
				locationRepo.On("GetAllRawLocations", ctx, db).
					Once().
					Return([]*domain.Location{
						{
							LocationID:        "location-org",
							PartnerInternalID: "loc_org",
							LocationType:      "org",
						},
						{
							LocationID:              "location-brand",
							PartnerInternalID:       "loc_brand",
							LocationType:            "brand",
							PartnerInternalParentID: "",
						},
						{
							LocationID:              "location-center",
							PartnerInternalID:       "loc_center",
							LocationType:            "center",
							PartnerInternalParentID: "loc_brand",
						},
					}, nil)

				tx.On("Rollback", ctx).Once().Return(nil)
			},
		},
		{
			name: "valid file with valid values should be imported",
			ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			req: &mpb.ImportLocationV2Request{
				Payload: []byte(`partner_internal_id,name,location_type,partner_internal_parent_id
				partner B,Location 1,brand,
				partner C,Location 2,center,partner B
				partner D,Location 3,area,partner C`),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db.On("Begin", ctx).Once().Return(tx, nil)
				locationRepo.On("GetLocationByLocationTypeName", ctx, db, "org").
					Return([]*domain.Location{
						{
							LocationID:        "location-id-0",
							LocationType:      "location-type-0",
							PartnerInternalID: "partner A",
						},
					}, nil).
					Once()
				locationTypeRepo.On("RetrieveLocationTypes", ctx, db).
					Once().
					Return([]*domain.LocationType{
						{
							LocationTypeID: "location-type-0",
							Name:           "org",
							Level:          0,
						},
						{
							LocationTypeID: "location-type-1",
							Name:           "brand",
							Level:          1,
						},
						{
							LocationTypeID: "location-type-2",
							Name:           "center",
							Level:          2,
						},
						{
							LocationTypeID: "location-type-3",
							Name:           "area",
							Level:          3,
						},
					}, nil)
				locationRepo.On("GetAllRawLocations", ctx, db).
					Once().
					Return([]*domain.Location{
						{
							LocationID:        "location-id-0",
							PartnerInternalID: "",
							LocationType:      "location-type-0",
						},
					}, nil)
				expectedLocations := []*domain.Location{
					{
						PartnerInternalID:       "partner B",
						PartnerInternalParentID: "",
						Name:                    "Location 1",
						IsArchived:              false,
						LocationType:            "location-type-1",
					},
					{
						PartnerInternalID:       "partner C",
						PartnerInternalParentID: "partner B",
						Name:                    "Location 2",
						IsArchived:              false,
						LocationType:            "location-type-2",
					},
					{
						PartnerInternalID:       "partner D",
						PartnerInternalParentID: "partner C",
						Name:                    "Location 3",
						IsArchived:              false,
						LocationType:            "location-type-3",
					},
				}
				locationRepo.On("UpsertLocations", ctx, db, mock.Anything).
					Once().
					Return(nil).
					Run(func(args mock.Arguments) {
						actualLocations := args.Get(2).([]*domain.Location)
						require.Equal(t, len(expectedLocations), len(actualLocations))
						locMap := make(map[string]*domain.Location)
						for _, v := range actualLocations {
							locMap[v.PartnerInternalID] = v
						}
						for i, expected := range expectedLocations {
							actual := actualLocations[i]
							// assert expected location
							require.Equal(t, actual.Name, expected.Name)
							require.Equal(t, actual.LocationType, expected.LocationType)
							require.Equal(t, actual.PartnerInternalParentID, expected.PartnerInternalParentID)
							require.Equal(t, actual.PartnerInternalID, expected.PartnerInternalID)

							// ensures we assign the correct parent location id from partner ID
							v, ok := locMap[actual.PartnerInternalParentID]
							if ok {
								require.Equal(t, v.LocationID, actual.ParentLocationID)
							} else {
								// parent is root
								require.Equal(t, "location-id-0", actual.ParentLocationID)

							}
						}
					})
				locationRepo.On("UpdateAccessPath", ctx, db, mock.Anything).
					Once().
					Return(nil)
				tx.On("Commit", ctx).Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.setup != nil {
				testCase.setup(testCase.ctx)
			}
			resp, err := s.ImportLocationV2(testCase.ctx, testCase.req.(*mpb.ImportLocationV2Request))
			if testCase.expectedErr != nil {
				assert.Nil(t, resp)
				if testCase.expectedErrModel != nil {
					utils.AssertBadRequestErrorModel(t, testCase.expectedErrModel, err)
				} else {
					assert.Equal(t, testCase.expectedErr, err)
				}
			} else {
				assert.Equal(t, nil, err)
				assert.NotNil(t, resp)
				mock.AssertExpectationsForObjects(t, locationTypeRepo)
				mock.AssertExpectationsForObjects(t, locationRepo)
			}
		})
	}
}
