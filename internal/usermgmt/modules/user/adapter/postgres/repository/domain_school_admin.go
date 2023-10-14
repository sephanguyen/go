package repository

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainSchoolAdminRepo struct {
	UserRepo            userRepo
	LegacyUserGroupRepo legacyUserGroupRepo
}

func NewDomainSchoolAdminRepo() *DomainSchoolAdminRepo {
	return &DomainSchoolAdminRepo{
		UserRepo:            &DomainUserRepo{},
		LegacyUserGroupRepo: &LegacyUserGroupRepo{},
	}
}

type userRepo interface {
	create(ctx context.Context, db database.QueryExecer, userToCreate entity.User) error
	UpsertMultiple(ctx context.Context, db database.QueryExecer, isEnableUsername bool, usersToCreate ...entity.User) error
}

type legacyUserGroupRepo interface {
	createMultiple(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error
}

type SchoolAdmin struct {
	entity.DomainSchoolAdmin

	// These attributes belong to postgres database context
	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func (e *SchoolAdmin) FieldMap() ([]string, []interface{}) {
	return []string{
			"school_admin_id", "school_id", "updated_at", "created_at", "resource_path",
		}, []interface{}{
			e.UserID().Ptr(), e.SchoolID().Ptr(), e.UpdatedAt.Ptr(), e.CreatedAt.Ptr(), e.OrganizationID().Ptr(),
		}
}

func (e *SchoolAdmin) TableName() string {
	return "school_admins"
}

func (repo *DomainSchoolAdminRepo) Create(ctx context.Context, db database.QueryExecer, schoolAdminToCreate aggregate.DomainSchoolAdmin) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainSchoolAdminRepo.Create")
	defer span.End()

	err := repo.UserRepo.create(ctx, db, schoolAdminToCreate)
	if err != nil {
		return err
	}

	err = repo.LegacyUserGroupRepo.createMultiple(ctx, db, schoolAdminToCreate.LegacyUserGroups...)
	if err != nil {
		return err
	}

	now := field.NewTime(time.Now())

	databaseSchoolAdminToCreate := &SchoolAdmin{
		DomainSchoolAdmin: schoolAdminToCreate,
		UpdatedAt:         now,
		CreatedAt:         now,
		DeletedAt:         now,
	}

	cmdTag, err := database.Insert(ctx, databaseSchoolAdminToCreate, db.Exec)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() != 1 {
		return ErrNoRowAffected
	}

	return nil
}
