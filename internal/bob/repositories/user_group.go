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

type UserGroupRepo struct{}

const userGroupFindStmtTpl = `SELECT %s
	FROM users_groups
	WHERE user_id = $1`

// Find finds all user groups that match userID
func (r *UserGroupRepo) Find(ctx context.Context, db database.QueryExecer, userID pgtype.Text) ([]*entities.UserGroup, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.Find")
	defer span.End()

	e := &entities.UserGroup{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf(userGroupFindStmtTpl, strings.Join(fields, ","))
	results := make(entities.UserGroups, 0, 10)
	if err := database.Select(ctx, db, query, &userID).ScanAll(&results); err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return results, nil
}

func (r *UserGroupRepo) Upsert(ctx context.Context, db database.QueryExecer, e *entities.UserGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.Upsert")
	defer span.End()

	now := time.Now()

	err := multierr.Combine(
		e.CreatedAt.Set(now),
		e.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set userGroup: %w", err)
	}

	fields, values := e.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fields))

	stmt := "INSERT INTO " + e.TableName() + " (" + strings.Join(fields, ",") + ") VALUES (" + placeHolders + ") ON CONFLICT (user_id,group_id) DO UPDATE SET updated_at = $6, is_origin = $3, status = $4;"

	cmd, err := db.Exec(ctx, stmt, values...)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmd.RowsAffected() == 0 {
		return errors.New("cannot upsert userGroup")
	}

	return nil
}

func (r *UserGroupRepo) UpdateStatus(ctx context.Context, db database.QueryExecer, userID, status pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.UpdateStatus")
	defer span.End()

	sql := `UPDATE users_groups SET updated_at = NOW(), status = $2 WHERE user_id = $1`
	cmdTag, err := db.Exec(ctx, sql, &userID, &status)
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("not found any records")
	}

	return nil
}

func (r *UserGroupRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, userGroups []*entities.UserGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.UserGroup) {
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

	for _, u := range userGroups {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(userGroups); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if ct.RowsAffected() != 1 {
			return fmt.Errorf("userGroup not inserted")
		}
	}

	return nil
}
