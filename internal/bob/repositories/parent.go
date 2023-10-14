package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ParentRepo provides method to work with parent entity
type ParentRepo struct{}

func (r *ParentRepo) Create(ctx context.Context, db database.QueryExecer, parent *entities.Parent) error {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.Create")
	defer span.End()

	now := time.Now()

	err := multierr.Combine(
		parent.ID.Set(parent.ID.String),
		parent.UpdatedAt.Set(now),
		parent.CreatedAt.Set(now),
		parent.User.ID.Set(parent.ID.String),
		parent.User.Group.Set(entities.UserGroupParent),
		parent.User.UpdatedAt.Set(now),
		parent.User.CreatedAt.Set(now),
		parent.User.DeviceToken.Set(nil),
		parent.User.AllowNotification.Set(true),
	)
	if err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	var userID pgtype.Text
	if err := database.InsertReturning(ctx, &parent.User, db, "user_id", &userID); err != nil {
		return errors.Wrap(err, "InsertReturning() user_id")
	}

	var parentID pgtype.Text
	if err := database.InsertReturning(ctx, parent, db, "parent_id", &parentID); err != nil {
		return errors.Wrap(err, "InsertReturning() parent_id")
	}

	group := &entities.UserGroup{}
	err = multierr.Combine(
		group.UserID.Set(parent.ID.String),
		group.GroupID.Set(entities.UserGroupParent),
		group.IsOrigin.Set(true),
		group.Status.Set(entities.UserGroupStatusActive),
		group.CreatedAt.Set(now),
		group.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set UserGroup: %w", err)
	}

	cmdTag, err := database.Insert(ctx, group, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert UserGroup: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("%d RowsAffected: %w", cmdTag.RowsAffected(), ErrUnAffected)
	}

	return nil
}

func (r *ParentRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, parents []*entities.Parent) error {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.Parent) {
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

func (r *ParentRepo) GetByIds(ctx context.Context, db database.QueryExecer, parentIds pgtype.TextArray) (entities.Parents, error) {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.GetById")
	defer span.End()

	parentEnt := &entities.Parent{}
	parentFieldNames := database.GetFieldNames(parentEnt)

	userEnt := &entities.User{}
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
		return entities.Parents{}, nil
	default:
		return nil, fmt.Errorf("db.Query: %w", err)
	}

	defer rows.Close()

	parents := make(entities.Parents, 0)
	for rows.Next() {
		parentEnt := &entities.Parent{}
		parentFieldValues := database.GetScanFields(parentEnt, parentFieldNames)

		userEnt := &entities.User{}
		userFieldValues := database.GetScanFields(userEnt, userFieldNames)

		fieldValues := make([]interface{}, 0, len(parentFieldValues)+len(userFieldValues))
		fieldValues = append(fieldValues, parentFieldValues...)
		fieldValues = append(fieldValues, userFieldValues...)

		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		parentEnt.User = *userEnt
		parents = append(parents, parentEnt)
	}

	return parents, nil
}
