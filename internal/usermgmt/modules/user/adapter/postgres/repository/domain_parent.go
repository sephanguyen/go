package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	db_usermgmt "github.com/manabie-com/backend/internal/usermgmt/pkg/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type DomainParentRepo struct {
	UserRepo            userRepo
	UserGroupMemberRepo userGroupMemberRepo
	LegacyUserGroupRepo legacyUserGroupRepo
	UserAccessPathRepo  userAccessPathRepo
}

type Parent struct {
	ID                 field.String
	SchoolIDAttr       field.Int32
	OrganizationIDAttr field.String
	CreatedAt          field.Time
	UpdatedAt          field.Time
	DeletedAt          field.Time
}

func NewParent(parent entity.DomainParent) *Parent {
	now := field.NewTime(time.Now())
	return &Parent{
		ID:                 parent.UserID(),
		SchoolIDAttr:       parent.SchoolID(),
		OrganizationIDAttr: parent.OrganizationID(),
		CreatedAt:          now,
		UpdatedAt:          now,
		DeletedAt:          field.NewNullTime(),
	}
}

func (parent *Parent) FieldMap() ([]string, []interface{}) {
	return []string{
			"parent_id",
			"school_id",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&parent.ID,
			&parent.SchoolIDAttr,
			&parent.UpdatedAt,
			&parent.CreatedAt,
			&parent.DeletedAt,
			&parent.OrganizationIDAttr,
		}
}

func (parent *Parent) TableName() string {
	return "parents"
}

func (repo *DomainParentRepo) GetUsersByExternalUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.Users, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainParentRepo.GetByExternalUserIDs")
	defer span.End()

	stmt := `
		SELECT users.%s
		FROM %s
		JOIN parents ON
		        parents.parent_id = users.user_id
		    AND parents.deleted_at IS NULL
		WHERE
			    users.user_external_id = ANY($1)
			AND users.deleted_at IS NULL
	`

	user, err := NewUser(entity.EmptyUser{})
	if err != nil {
		return nil, InternalError{RawError: err}
	}

	fieldNames, _ := user.FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ", users."),
		user.TableName(),
	)

	rows, err := db.Query(ctx, stmt, database.TextArray(userIDs))
	if err != nil {
		return nil, InternalError{
			RawError: errors.Wrap(err, "db.Query"),
		}
	}

	defer rows.Close()

	result := make(entity.Users, 0, len(userIDs))
	for rows.Next() {
		item, err := NewUser(entity.EmptyUser{})
		if err != nil {
			return nil, InternalError{RawError: err}
		}

		_, fieldValues := item.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			return nil, InternalError{
				RawError: fmt.Errorf("rows.Scan: %w", err),
			}
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return nil, InternalError{RawError: err}
	}

	return result, nil
}

func (repo *DomainParentRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, isEnableUsername bool, parentsToUpsert ...aggregate.DomainParent) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainParentRepo.UpsertMultiple")
	defer span.End()
	usersToUpsert := entity.Users{}
	userGroupMembers := entity.DomainUserGroupMembers{}
	legacyUserGroups := entity.LegacyUserGroups{}
	userAccessPaths := entity.DomainUserAccessPaths{}

	for _, parent := range parentsToUpsert {
		usersToUpsert = append(usersToUpsert, parent)
		userGroupMembers = append(userGroupMembers, parent.UserGroupMembers...)
		userAccessPaths = append(userAccessPaths, parent.UserAccessPaths...)
		legacyUserGroups = append(legacyUserGroups, parent.LegacyUserGroups...)
	}

	err := repo.UserAccessPathRepo.SoftDeleteByUserIDs(ctx, db, usersToUpsert.UserIDs())
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserAccessPathRepo.SoftDeleteByUserIDs")}
	}

	err = repo.UserAccessPathRepo.UpsertMultiple(ctx, db, userAccessPaths...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserAccessPathRepo.upsertMultiple")}
	}

	err = repo.UserRepo.UpsertMultiple(ctx, db, isEnableUsername, usersToUpsert...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserRepo.UpsertMultiple")}
	}
	err = repo.LegacyUserGroupRepo.createMultiple(ctx, db, legacyUserGroups...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.LegacyUserGroupRepo.createMultiple")}
	}
	err = repo.UserGroupMemberRepo.CreateMultiple(ctx, db, userGroupMembers...)
	if err != nil {
		return InternalError{RawError: errors.Wrap(err, "repo.UserGroupMemberRepo.CreateMultiple")}
	}
	batch := &pgx.Batch{}

	queueFn := func(b *pgx.Batch, parent *Parent) {
		fields, values := parent.FieldMap()

		insertPlaceHolders := database.GeneratePlaceholders(len(fields))
		updatePlaceHolders := db_usermgmt.GenerateUpdatePlaceholders(fields)
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT parents__parent_id_pk DO UPDATE SET %s",
			parent.TableName(),
			strings.Join(fields, ","),
			insertPlaceHolders,
			updatePlaceHolders,
		)

		b.Queue(stmt, values...)
	}
	for _, parentToUpsert := range parentsToUpsert {
		databaseStudentToUpsert := NewParent(parentToUpsert)

		queueFn(batch, databaseStudentToUpsert)
	}
	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()
	for i := 0; i < len(parentsToUpsert); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return InternalError{RawError: errors.Wrap(err, "batchResults.Exec")}
		}

		if cmdTag.RowsAffected() != 1 {
			return InternalError{RawError: errors.Errorf("parent was not inserted")}
		}
	}

	return nil
}
