package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDomainUserRepo struct {
	createFn         func(ctx context.Context, db database.QueryExecer, userToCreate entity.User) error
	upsertMultipleFn func(ctx context.Context, db database.QueryExecer, isEnableUsername bool, usersToCreate ...entity.User) error
}

func (mockDomainUserRepo mockDomainUserRepo) create(ctx context.Context, db database.QueryExecer, userToCreate entity.User) error {
	return mockDomainUserRepo.createFn(ctx, db, userToCreate)
}

func (mockDomainUserRepo mockDomainUserRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, isEnableUsername bool, usersToCreate ...entity.User) error {
	return mockDomainUserRepo.upsertMultipleFn(ctx, db, isEnableUsername, usersToCreate...)
}

type mockLegacyUserGroupRepo struct {
	createMultipleFn func(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error
}

func (mockLegacyUserGroupRepo mockLegacyUserGroupRepo) createMultiple(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error {
	return mockLegacyUserGroupRepo.createMultipleFn(ctx, db, legacyUserGroups...)
}

func TestDomainSchoolAdminRepo_Create(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Parallel()

	t.Run("happy case", func(t *testing.T) {
		domainSchoolAdminRepo := DomainSchoolAdminRepo{
			UserRepo: mockDomainUserRepo{
				createFn: func(ctx context.Context, db database.QueryExecer, userToCreate entity.User) error {
					return nil
				},
			},
			LegacyUserGroupRepo: mockLegacyUserGroupRepo{
				createMultipleFn: func(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error {
					return nil
				},
			},
		}

		mockDB := testutil.NewMockDB()

		domainSchoolAdmin := aggregate.DomainSchoolAdmin{
			DomainSchoolAdmin: entity.NullDomainSchoolAdmin{},
			LegacyUserGroups:  entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
		}

		repoSchoolAdmin := &SchoolAdmin{
			DomainSchoolAdmin: domainSchoolAdmin,
		}

		_, userValues := repoSchoolAdmin.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)

		cmdTag := pgconn.CommandTag(`1`)
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := domainSchoolAdminRepo.Create(ctx, mockDB.DB, domainSchoolAdmin)
		assert.Nil(t, err)
	})

	t.Run("failed to create", func(t *testing.T) {
		domainSchoolAdminRepo := DomainSchoolAdminRepo{
			UserRepo: mockDomainUserRepo{
				createFn: func(ctx context.Context, db database.QueryExecer, userToCreate entity.User) error {
					return nil
				},
			},
			LegacyUserGroupRepo: mockLegacyUserGroupRepo{
				createMultipleFn: func(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error {
					return nil
				},
			},
		}

		mockDB := testutil.NewMockDB()

		domainSchoolAdmin := aggregate.DomainSchoolAdmin{
			DomainSchoolAdmin: entity.NullDomainSchoolAdmin{},
			LegacyUserGroups:  entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
		}

		repoSchoolAdmin := &SchoolAdmin{
			DomainSchoolAdmin: domainSchoolAdmin,
		}

		_, userValues := repoSchoolAdmin.FieldMap()
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userValues))...)

		mockDB.DB.On("Exec", args...).Return(nil, puddle.ErrClosedPool)

		err := domainSchoolAdminRepo.Create(ctx, mockDB.DB, domainSchoolAdmin)
		assert.Equal(t, puddle.ErrClosedPool, err)
	})
}
