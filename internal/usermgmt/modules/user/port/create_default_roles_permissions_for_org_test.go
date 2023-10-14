package port

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/location/infrastructure/repo"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrganizationEvent_Subscribe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		init    func(t *testing.T) *CreateOrganizationEvent
		inspect func(r *CreateOrganizationEvent, t *testing.T)

		wantErr    bool
		inspectErr func(err error, t *testing.T)
	}{
		{
			name:    "happy case",
			wantErr: false,
			init: func(t *testing.T) *CreateOrganizationEvent {
				jsm := &mock_nats.JetStreamManagement{}
				jsm.On("QueueSubscribe", constants.SubjectSyncLocationUpserted, constants.QueueSyncLocationUpsertedOrgCreation, mock.Anything, mock.Anything).Once().Return(nil, nil)
				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(nil, jsm, nil, nil, nil, nil)
				return newCreateOrganizationEventHandler
			},
		},
		{
			name:    "error when subscribing event",
			wantErr: true,
			inspectErr: func(err error, t *testing.T) {
				assert.EqualError(t, err, errors.Wrapf(fmt.Errorf("error"), "c.JSM.QueueSubscribe: %s", constants.QueueSyncLocationUpsertedOrgCreation).Error())
			},
			init: func(t *testing.T) *CreateOrganizationEvent {
				jsm := &mock_nats.JetStreamManagement{}
				jsm.On("QueueSubscribe", constants.SubjectSyncLocationUpserted, constants.QueueSyncLocationUpsertedOrgCreation, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(nil, jsm, nil, nil, nil, nil)
				return newCreateOrganizationEventHandler
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receiver := tt.init(t)
			err := receiver.Subscribe()

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("createOrganizationEvent.Subscribe error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}

func TestCreateOrganizationEvent_createOrganizationEventHandler(t *testing.T) {
	locationID := idutil.ULIDNow()
	type args struct {
		ctx  context.Context
		data []byte
	}
	tests := []struct {
		name    string
		init    func(t *testing.T) *CreateOrganizationEvent
		inspect func(r *CreateOrganizationEvent, t *testing.T)

		args func(t *testing.T) args

		wantReturn bool
		wantErr    bool
		inspectErr func(err error, t *testing.T)
	}{
		{
			name:       "happy case",
			wantReturn: true,
			wantErr:    false,
			init: func(t *testing.T) *CreateOrganizationEvent {
				zapLogger := logger.NewZapLogger("", true)
				tx := new(mock_database.Tx)
				locationRepo := new(mock_repo.MockLocationRepo)
				roleRepo := new(mock_repositories.MockRoleRepo)
				permissionRepo := new(mock_repositories.MockPermissionRepo)
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&domain.Location{ResourcePath: fmt.Sprint(constants.ManabieSchool)}, nil)

				tx.On("Begin", mock.Anything).Return(tx, nil)
				roleRepo.On("Create", mock.Anything, tx, mock.Anything).Return(nil)
				permissionRepo.On("CreateBatch", mock.Anything, tx, mock.Anything).Return(nil)
				roleRepo.On("UpsertPermission", mock.Anything, tx, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)

				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(zapLogger, nil, tx, roleRepo, permissionRepo, locationRepo)
				return newCreateOrganizationEventHandler
			},
			args: func(t *testing.T) args {
				eventLocation := &npb.EventSyncLocation{
					Locations: []*npb.EventSyncLocation_Location{
						{LocationId: idutil.ULIDNow()},
					},
				}
				data, _ := proto.Marshal(eventLocation)
				return args{
					ctx:  context.Background(),
					data: data,
				}
			},
		},
		{
			name:       "happy case: empty location",
			wantReturn: true,
			init: func(t *testing.T) *CreateOrganizationEvent {
				locationRepo := &mock_repo.MockLocationRepo{}
				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(nil, nil, nil, nil, nil, locationRepo)
				return newCreateOrganizationEventHandler
			},
			args: func(t *testing.T) args {
				event := &npb.EventSyncLocation{}
				data, _ := proto.Marshal(event)
				return args{
					ctx:  context.Background(),
					data: data,
				}
			},
		},
		{
			name:       "continue process althought one location got stuck at create role",
			wantReturn: true,
			wantErr:    false,
			init: func(t *testing.T) *CreateOrganizationEvent {
				zapLogger := logger.NewZapLogger("", true)
				tx := new(mock_database.Tx)
				locationRepo := new(mock_repo.MockLocationRepo)
				roleRepo := new(mock_repositories.MockRoleRepo)
				permissionRepo := new(mock_repositories.MockPermissionRepo)
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, mock.Anything).Return(&domain.Location{ResourcePath: fmt.Sprint(constants.ManabieSchool)}, nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				roleRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Return(tx, nil)
				roleRepo.On("Create", mock.Anything, tx, mock.Anything).Return(nil)
				permissionRepo.On("CreateBatch", mock.Anything, tx, mock.Anything).Return(nil)
				roleRepo.On("UpsertPermission", mock.Anything, tx, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)

				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(zapLogger, nil, tx, roleRepo, permissionRepo, locationRepo)
				return newCreateOrganizationEventHandler
			},
			args: func(t *testing.T) args {
				eventLocation := &npb.EventSyncLocation{
					Locations: []*npb.EventSyncLocation_Location{
						{LocationId: idutil.ULIDNow()},
						{LocationId: idutil.ULIDNow()},
						{LocationId: idutil.ULIDNow()},
					},
				}
				data, _ := proto.Marshal(eventLocation)
				return args{
					ctx:  context.Background(),
					data: data,
				}
			},
		},
		{
			name:       "continue process althought one location got stuck at create permissions",
			wantReturn: true,
			wantErr:    false,
			init: func(t *testing.T) *CreateOrganizationEvent {
				zapLogger := logger.NewZapLogger("", true)
				tx := new(mock_database.Tx)
				locationRepo := new(mock_repo.MockLocationRepo)
				roleRepo := new(mock_repositories.MockRoleRepo)
				permissionRepo := new(mock_repositories.MockPermissionRepo)
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&domain.Location{ResourcePath: fmt.Sprint(constants.ManabieSchool)}, nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				roleRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				permissionRepo.On("CreateBatch", mock.Anything, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Return(tx, nil)
				roleRepo.On("Create", mock.Anything, tx, mock.Anything).Return(nil)
				permissionRepo.On("CreateBatch", mock.Anything, tx, mock.Anything).Return(nil)
				roleRepo.On("UpsertPermission", mock.Anything, tx, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)

				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(zapLogger, nil, tx, roleRepo, permissionRepo, locationRepo)
				return newCreateOrganizationEventHandler
			},
			args: func(t *testing.T) args {
				eventLocation := &npb.EventSyncLocation{
					Locations: []*npb.EventSyncLocation_Location{
						{LocationId: idutil.ULIDNow()},
					},
				}
				data, _ := proto.Marshal(eventLocation)
				return args{
					ctx:  context.Background(),
					data: data,
				}
			},
		},
		{
			name:       "continue process althought one location got stuck at upsert permissions",
			wantReturn: true,
			wantErr:    false,
			init: func(t *testing.T) *CreateOrganizationEvent {
				zapLogger := logger.NewZapLogger("", true)
				tx := new(mock_database.Tx)
				locationRepo := new(mock_repo.MockLocationRepo)
				roleRepo := new(mock_repositories.MockRoleRepo)
				permissionRepo := new(mock_repositories.MockPermissionRepo)
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(&domain.Location{ResourcePath: fmt.Sprint(constants.ManabieSchool)}, nil)

				tx.On("Begin", mock.Anything).Once().Return(tx, nil)
				roleRepo.On("Create", mock.Anything, tx, mock.Anything).Once().Return(nil)
				permissionRepo.On("CreateBatch", mock.Anything, tx, mock.Anything).Once().Return(nil)
				roleRepo.On("UpsertPermission", mock.Anything, tx, mock.Anything).Once().Return(fmt.Errorf("error"))
				tx.On("Rollback", mock.Anything).Once().Return(nil)

				tx.On("Begin", mock.Anything).Return(tx, nil)
				roleRepo.On("Create", mock.Anything, tx, mock.Anything).Return(nil)
				permissionRepo.On("CreateBatch", mock.Anything, tx, mock.Anything).Return(nil)
				roleRepo.On("UpsertPermission", mock.Anything, tx, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)

				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(zapLogger, nil, tx, roleRepo, permissionRepo, locationRepo)
				return newCreateOrganizationEventHandler
			},
			args: func(t *testing.T) args {
				eventLocation := &npb.EventSyncLocation{
					Locations: []*npb.EventSyncLocation_Location{
						{LocationId: idutil.ULIDNow()},
					},
				}
				data, _ := proto.Marshal(eventLocation)
				return args{
					ctx:  context.Background(),
					data: data,
				}
			},
		},
		{
			name:       "internal error: error occur in locationRepo.GetLocationByID",
			wantReturn: false,
			wantErr:    true,
			inspectErr: func(err error, t *testing.T) {
				assert.EqualError(t, err, errors.Wrap(errors.Wrapf(fmt.Errorf("error"), "locationRepo.GetLocationByID: %s", locationID), "c.handleRolePermisionForLocation").Error())
			},
			init: func(t *testing.T) *CreateOrganizationEvent {
				locationRepo := new(mock_repo.MockLocationRepo)
				locationRepo.On("GetLocationByID", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))

				newCreateOrganizationEventHandler := NewCreateOrganizationEventHandler(nil, nil, nil, nil, nil, locationRepo)
				return newCreateOrganizationEventHandler
			},
			args: func(t *testing.T) args {
				eventLocation := &npb.EventSyncLocation{
					Locations: []*npb.EventSyncLocation_Location{
						{LocationId: locationID},
					},
				}
				data, _ := proto.Marshal(eventLocation)
				return args{
					ctx:  context.Background(),
					data: data,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tArgs := tt.args(t)

			receiver := tt.init(t)
			got1, err := receiver.createOrganizationEventHandler(tArgs.ctx, tArgs.data)

			if tt.inspect != nil {
				tt.inspect(receiver, t)
			}

			if !assert.Equal(t, got1, tt.wantReturn) {
				t.Errorf("createOrganizationEvent.createOrganizationEventHandler got1 = %v, want1: %v", got1, tt.wantReturn)
			}

			if (err != nil) != tt.wantErr {
				t.Fatalf("createOrganizationEvent.createOrganizationEventHandler error = %v, wantErr: %t", err, tt.wantErr)
			}

			if tt.inspectErr != nil {
				tt.inspectErr(err, t)
			}
		})
	}
}
