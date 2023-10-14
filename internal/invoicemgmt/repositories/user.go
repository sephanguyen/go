package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"github.com/jackc/pgx/v4"
)

type UserRepo struct{}

func (r *UserRepo) FindByID(ctx context.Context, db database.QueryExecer, userID string) (*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.FindByID")
	defer span.End()

	e := &entities.User{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", strings.Join(fields, ","), e.TableName())
	row := db.QueryRow(ctx, query, &userID)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return e, nil
}

func (r *UserRepo) FindUserWithEmailByEmailReference(ctx context.Context, db database.QueryExecer, emailReference string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.FindUserWithEmailByEmailReference")
	defer span.End()

	user := &entities.User{}
	var userID string

	query := fmt.Sprintf("SELECT user_id FROM %s WHERE email = $1 AND deleted_at IS NULL", user.TableName())

	err := db.QueryRow(ctx, query, &emailReference).Scan(&userID)

	switch err {
	case nil:
		return userID, nil
	case pgx.ErrNoRows:
		return "", nil
	default:
		return "", fmt.Errorf("err UserRepo FindUserWithEmailByEmailReference: %w", err)
	}
}

func (r *UserRepo) FindByUserExternalID(ctx context.Context, db database.QueryExecer, externalUserID string) (*entities.User, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepo.FindByID")
	defer span.End()

	e := &entities.User{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_external_id = $1", strings.Join(fields, ","), e.TableName())
	row := db.QueryRow(ctx, query, &externalUserID)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return e, nil
}
