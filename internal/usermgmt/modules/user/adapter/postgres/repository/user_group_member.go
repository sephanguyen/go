package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type UserGroupsMemberRepo struct{}

func (r *UserGroupsMemberRepo) GetByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entity.UserGroupMember, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.GetUserGroupMembers")
	defer span.End()

	fields, _ := new(entity.UserGroupMember).FieldMap()
	stmt := fmt.Sprintf(`SELECT ugm.%s FROM user_group_member ugm WHERE ugm.user_id = $1 AND ugm.deleted_at IS NULL`, strings.Join(fields, ", ugm."))

	userGroupMembers := entity.UserGroupMembers{}
	if err := database.Select(ctx, db, stmt, &userID).ScanAll(&userGroupMembers); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return userGroupMembers, nil
}

func (r *UserGroupsMemberRepo) UpsertBatch(ctx context.Context, db database.QueryExecer, userGroupsMembers []*entity.UserGroupMember) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupsMemberRepo.UpsertBatch")
	defer span.End()

	batch := &pgx.Batch{}
	if err := r.queueUpsert(ctx, batch, userGroupsMembers); err != nil {
		return fmt.Errorf("queueUpsert error: %w", err)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
	}

	return nil
}

func (r *UserGroupsMemberRepo) SoftDelete(ctx context.Context, db database.QueryExecer, userGroupsMembers []*entity.UserGroupMember) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupsMemberRepo.SoftDelete")
	defer span.End()

	userIDs := []string{}
	userGroupIDs := []string{}
	for _, ugm := range userGroupsMembers {
		userIDs = append(userIDs, ugm.UserID.String)
		userGroupIDs = append(userGroupIDs, ugm.UserGroupID.String)
	}

	query := `
	  UPDATE
	      user_group_member
	  SET deleted_at = now(),
	      updated_at = now()
	  WHERE
	      user_id = ANY($1) AND
	      user_group_id = ANY($2) AND
	      deleted_at IS NULL`
	_, err := db.Exec(ctx, query, &userIDs, &userGroupIDs)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserGroupsMemberRepo) queueUpsert(ctx context.Context, batch *pgx.Batch, userGroupsMembers []*entity.UserGroupMember) error {
	queue := func(btch *pgx.Batch, userGroupsMember *entity.UserGroupMember) {
		fieldNames := database.GetFieldNames(userGroupsMember)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		stmt := fmt.Sprintf(`
			INSERT INTO %s (%s) VALUES (%s)
			ON CONFLICT ON CONSTRAINT pk__user_group_member 
			DO UPDATE SET created_at = $3, updated_at = $4, deleted_at = NULL`,
			userGroupsMember.TableName(),
			strings.Join(fieldNames, ","),
			placeHolders,
		)

		btch.Queue(stmt, database.GetScanFields(userGroupsMember, fieldNames)...)
	}

	now := time.Now()
	for _, ugmEnt := range userGroupsMembers {
		if ugmEnt.UserGroupID.Status != pgtype.Present {
			continue
		}

		if ugmEnt.ResourcePath.Status == pgtype.Null {
			resourcePath := golibs.ResourcePathFromCtx(ctx)
			if err := ugmEnt.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		if err := multierr.Combine(
			ugmEnt.CreatedAt.Set(now),
			ugmEnt.UpdatedAt.Set(now),
		); err != nil {
			return err
		}

		queue(batch, ugmEnt)
	}

	return nil
}

func (r *UserGroupsMemberRepo) AssignWithUserGroup(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser, userGroupID pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupsMemberRepo.AssignWithUserGroup")
	defer span.End()

	queueFn := func(batch *pgx.Batch, userGroupMember *entity.UserGroupMember) {
		fields, values := userGroupMember.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := "INSERT INTO " + userGroupMember.TableName() + " (" + strings.Join(fields, ",") + ") VALUES (" + placeHolders + ");"

		batch.Queue(stmt, values...)
	}

	now := time.Now()
	batch := &pgx.Batch{}
	for index := range users {
		userGroupMember := new(entity.UserGroupMember)
		if err := multierr.Combine(
			userGroupMember.UserID.Set(users[index].ID),
			userGroupMember.UserGroupID.Set(userGroupID),
			userGroupMember.CreatedAt.Set(now),
			userGroupMember.UpdatedAt.Set(now),
			userGroupMember.DeletedAt.Set(nil),
			userGroupMember.ResourcePath.Set(users[index].ResourcePath),
		); err != nil {
			return fmt.Errorf("multierr.Combine: %w", err)
		}
		queueFn(batch, userGroupMember)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range users {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("user group member not inserted")
		}
	}

	return nil
}
