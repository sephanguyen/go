package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DomainUserGroupMemberRepo struct{}

type UserGroupMemberAttribute struct {
	UserID         field.String
	UserGroupID    field.String
	OrganizationID field.String
}

type UserGroupMember struct {
	UserGroupMemberAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func newDomainUserGroupMember(ugm entity.DomainUserGroupMember) *UserGroupMember {
	now := field.NewTime(time.Now())
	return &UserGroupMember{
		UserGroupMemberAttribute: UserGroupMemberAttribute{
			UserID:         ugm.UserID(),
			UserGroupID:    ugm.UserGroupID(),
			OrganizationID: ugm.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (ugm *UserGroupMember) UserID() field.String {
	return ugm.UserGroupMemberAttribute.UserID
}
func (ugm *UserGroupMember) UserGroupID() field.String {
	return ugm.UserGroupMemberAttribute.UserGroupID
}
func (ugm *UserGroupMember) OrganizationID() field.String {
	return ugm.UserGroupMemberAttribute.OrganizationID
}

func (*UserGroupMember) TableName() string {
	return "user_group_member"
}

func (ugm *UserGroupMember) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"user_group_id",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&ugm.UserGroupMemberAttribute.UserID,
			&ugm.UserGroupMemberAttribute.UserGroupID,
			&ugm.CreatedAt,
			&ugm.UpdatedAt,
			&ugm.UserGroupMemberAttribute.OrganizationID,
		}
}

func (repo *DomainUserGroupMemberRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, userGroupMembers ...entity.DomainUserGroupMember) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserGroupMemberRepo.createMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, userGroupMember *UserGroupMember) {
		fields, values := userGroupMember.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__user_group_member DO NOTHING",
			userGroupMember.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	for _, userGroupMember := range userGroupMembers {
		repoUserGroupMember := newDomainUserGroupMember(userGroupMember)

		queueFn(batch, repoUserGroupMember)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(userGroupMembers); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}
