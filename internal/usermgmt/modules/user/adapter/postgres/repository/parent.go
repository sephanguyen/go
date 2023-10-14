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
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ErrUnAffected for unexpected value returned by commandTag
var ErrUnAffected = errors.New("unexpected RowsAffected value")

// ParentRepo provides method to work with parent entity
type ParentRepo struct{}

func (r *ParentRepo) Create(ctx context.Context, db database.QueryExecer, parent *entity.Parent) error {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.Create")
	defer span.End()

	now := time.Now()

	err := multierr.Combine(
		parent.ID.Set(parent.ID.String),
		parent.UpdatedAt.Set(now),
		parent.CreatedAt.Set(now),
		parent.LegacyUser.ID.Set(parent.ID.String),
		parent.LegacyUser.Group.Set(entity.UserGroupParent),
		parent.LegacyUser.UpdatedAt.Set(now),
		parent.LegacyUser.CreatedAt.Set(now),
		parent.LegacyUser.DeviceToken.Set(nil),
		parent.LegacyUser.AllowNotification.Set(true),
		parent.LegacyUser.UserRole.Set(string(constant.UserRoleParent)),
	)
	if err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	if parent.LegacyUser.ResourcePath.Status == pgtype.Null {
		resourcePath := golibs.ResourcePathFromCtx(ctx)
		if err := parent.LegacyUser.ResourcePath.Set(resourcePath); err != nil {
			return err
		}
	}

	cmdTag, err := database.Insert(ctx, &parent.LegacyUser, db.Exec)
	if err != nil {
		return errors.Wrap(err, "Insert() user_id")
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	cmdTag, err = database.Insert(ctx, parent, db.Exec)
	if err != nil {
		return errors.Wrap(err, "Insert() parent_id")
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	group := &entity.UserGroup{}
	err = multierr.Combine(
		group.UserID.Set(parent.ID.String),
		group.GroupID.Set(entity.UserGroupParent),
		group.IsOrigin.Set(true),
		group.Status.Set(entity.UserGroupStatusActive),
		group.CreatedAt.Set(now),
		group.UpdatedAt.Set(now),
		group.ResourcePath.Set(parent.LegacyUser.ResourcePath),
	)
	if err != nil {
		return fmt.Errorf("err set UserGroup: %w", err)
	}

	cmdTag, err = database.Insert(ctx, group, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert UserGroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	return nil
}

func (r *ParentRepo) GetByID(ctx context.Context, db database.QueryExecer, parentID pgtype.Text) (*entity.Parent, error) {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.GetByID")
	defer span.End()

	parentEnt := &entity.Parent{}
	fields := database.GetFieldNames(parentEnt)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE parent_id = $1 and deleted_at IS NULL", strings.Join(fields, ","), parentEnt.TableName())
	row := db.QueryRow(ctx, query, &parentID)
	if err := row.Scan(database.GetScanFields(parentEnt, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return parentEnt, nil
}

func (r *ParentRepo) GetByIds(ctx context.Context, db database.QueryExecer, parentIds pgtype.TextArray) (entity.Parents, error) {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.GetById")
	defer span.End()

	parentEnt := &entity.Parent{}
	parentFieldNames := database.GetFieldNames(parentEnt)

	userEnt := &entity.LegacyUser{}
	userFieldNames := database.GetFieldNames(userEnt)

	stmt :=
		`
		SELECT 
			parents.%s, users.%s
		FROM 
			%s 
		JOIN
			%s ON parents.parent_id = users.user_id
		WHERE 
			parents.parent_id = ANY($1) AND parents.deleted_at IS NULL
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(parentFieldNames, ", parents."),
		strings.Join(userFieldNames, ", users."),
		parentEnt.TableName(),
		userEnt.TableName(),
	)

	rows, err := db.Query(ctx, stmt, &parentIds)
	switch err {
	case nil:
		break
	case pgx.ErrNoRows:
		return entity.Parents{}, nil
	default:
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	defer rows.Close()

	parents := make(entity.Parents, 0)
	for rows.Next() {
		parentEnt := &entity.Parent{}
		parentFieldValues := database.GetScanFields(parentEnt, parentFieldNames)

		userEnt := &entity.LegacyUser{}
		userFieldValues := database.GetScanFields(userEnt, userFieldNames)

		fieldValues := make([]interface{}, 0, len(parentFieldValues)+len(userFieldValues))
		fieldValues = append(fieldValues, parentFieldValues...)
		fieldValues = append(fieldValues, userFieldValues...)

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		parentEnt.LegacyUser = *userEnt
		parents = append(parents, parentEnt)
	}

	return parents, nil
}

func (r *ParentRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, parents []*entity.Parent) error {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entity.Parent) {
		fields, values := u.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			u.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	b := &pgx.Batch{}
	now := time.Now()

	for _, u := range parents {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		u.UserRole = database.Text(string(constant.UserRoleParent))
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(parents); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("parent not inserted")
		}
	}

	return nil
}
