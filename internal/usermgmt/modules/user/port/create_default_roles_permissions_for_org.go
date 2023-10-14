package port

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var RoleWithPermissionForOrg = map[string][]string{
	constant.RoleHQStaff:       {constant.MasterLocationRead},
	constant.RoleSchoolAdmin:   {constant.MasterLocationRead},
	constant.RoleCentreManager: {constant.MasterLocationRead},
	constant.RoleCentreStaff:   {constant.MasterLocationRead},
	constant.RoleCentreLead:    {constant.MasterLocationRead},
	constant.RoleTeacherLead:   {constant.MasterLocationRead},
	constant.RoleTeacher:       {constant.MasterLocationRead},
	constant.RoleParent:        {constant.MasterLocationRead},
	constant.RoleStudent:       {constant.MasterLocationRead},
}

type RoleRepo interface {
	Create(ctx context.Context, db database.Ext, role *entity.Role) error
	UpsertPermission(ctx context.Context, db database.Ext, permissionRoles []*entity.PermissionRole) error
}

type PermissionRepo interface {
	CreateBatch(ctx context.Context, db database.Ext, permissions []*entity.Permission) error
}

type LocationRepo interface {
	GetLocationByID(ctx context.Context, db database.Ext, id string) (*domain.Location, error)
}

type CreateOrganizationEvent struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	DB     database.Ext

	RoleRepo       RoleRepo
	PermissionRepo PermissionRepo
	LocationRepo   LocationRepo
}

func NewCreateOrganizationEventHandler(
	logger *zap.Logger,
	jsm nats.JetStreamManagement,
	db database.Ext,
	roleRepo RoleRepo,
	permissionRole PermissionRepo,
	locationRepo LocationRepo,
) *CreateOrganizationEvent {
	return &CreateOrganizationEvent{
		Logger:         logger,
		JSM:            jsm,
		DB:             db,
		RoleRepo:       roleRepo,
		PermissionRepo: permissionRole,
		LocationRepo:   locationRepo,
	}
}

func (c *CreateOrganizationEvent) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.AckWait(constant.JSMAckWait),
			nats.MaxDeliver(constant.JSMMaxDeliver),
			nats.Bind(constants.StreamSyncLocationUpserted, constants.DurableSyncLocationUpsertedOrgCreation),
			nats.DeliverSubject(constants.DeliverSyncLocationUpsertedOrgCreation),
			// nats.DeliverNew(),
		},
	}

	if _, err := c.JSM.QueueSubscribe(
		constants.SubjectSyncLocationUpserted,
		constants.QueueSyncLocationUpsertedOrgCreation,
		option,
		c.createOrganizationEventHandler,
	); err != nil {
		return errors.Wrapf(err, "c.JSM.QueueSubscribe: %s", constants.QueueSyncLocationUpsertedOrgCreation)
	}

	return nil
}

func (c *CreateOrganizationEvent) createOrganizationEventHandler(ctx context.Context, data []byte) (bool, error) {
	eventLocation := &npb.EventSyncLocation{}
	if err := proto.Unmarshal(data, eventLocation); err != nil {
		return false, errors.Wrap(err, "proto.Unmarshal: eventLocation")
	}

	for _, location := range eventLocation.Locations {
		if err := c.handleRolePermissionForLocation(ctx, location); err != nil {
			return false, errors.Wrap(err, "c.handleRolePermisionForLocation")
		}
	}
	return true, nil
}

func (c *CreateOrganizationEvent) handleRolePermissionForLocation(ctx context.Context, locationEvent *npb.EventSyncLocation_Location) error {
	location, err := c.LocationRepo.GetLocationByID(ctx, c.DB, locationEvent.LocationId)
	if err != nil {
		return errors.Wrapf(err, "locationRepo.GetLocationByID: %s", locationEvent.LocationId)
	}

	for role, permissions := range RoleWithPermissionForOrg {
		if err := database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
			// create role
			roleEnt := newRoleEntity(role, location.ResourcePath)
			if err := c.RoleRepo.Create(ctx, tx, roleEnt); err != nil {
				return errors.Wrap(err, "roleRepo.Create")
			}

			// create permission
			permissionEntities := newPermissionEntities(permissions, location.ResourcePath)
			if err := c.PermissionRepo.CreateBatch(ctx, tx, permissionEntities); err != nil {
				return errors.Wrap(err, "permissionRepo.CreateBatch")
			}

			// assign permistion to role
			permissionRoleEntities := newPermissionRoleEntities(roleEnt, permissionEntities)
			if err := c.RoleRepo.UpsertPermission(ctx, tx, permissionRoleEntities); err != nil {
				return errors.Wrap(err, "roleRepo.UpsertPermission")
			}

			return nil
		}); err != nil {
			c.Logger.Error(errors.Wrapf(err, "database.ExecInTx: %s", role).Error())
		}
	}
	return nil
}

func newRoleEntity(roleName, resourcePath string) *entity.Role {
	role := new(entity.Role)
	database.AllNullEntity(role)

	role.RoleID = database.Text(idutil.ULIDNow())
	role.RoleName = database.Text(roleName)
	role.ResourcePath = database.Text(resourcePath)
	role.IsSystem = database.Bool(true)

	return role
}

func newPermissionEntities(permissionNames []string, resourcePath string) []*entity.Permission {
	permissions := make([]*entity.Permission, 0, len(permissionNames))
	for _, permissionName := range permissionNames {
		permission := new(entity.Permission)

		database.AllNullEntity(permission)
		permission.PermissionID = database.Text(idutil.ULIDNow())
		permission.PermissionName = database.Text(permissionName)
		permission.ResourcePath = database.Text(resourcePath)

		permissions = append(permissions, permission)
	}
	return permissions
}

func newPermissionRoleEntities(role *entity.Role, permissions []*entity.Permission) []*entity.PermissionRole {
	permissionRoles := make([]*entity.PermissionRole, 0, len(permissions))
	for _, permission := range permissions {
		permissionRole := new(entity.PermissionRole)

		database.AllNullEntity(permissionRole)
		permissionRole.PermissionID = permission.PermissionID
		permissionRole.RoleID = role.RoleID
		permissionRole.ResourcePath = role.ResourcePath

		permissionRoles = append(permissionRoles, permissionRole)
	}
	return permissionRoles
}
