package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type AppleUserRepo struct{}

func (r *AppleUserRepo) Create(ctx context.Context, db database.QueryExecer, user *entity.AppleUser) error {
	ctx, span := interceptors.StartSpan(ctx, "AppleUserRepo.Create")
	defer span.End()

	now := timeutil.Now()
	err := multierr.Combine(
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("create time now: %w", err)
	}

	if _, err := database.Insert(ctx, user, db.Exec); err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

func (r *AppleUserRepo) Get(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entity.AppleUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "AppleUserRepo.Get")
	defer span.End()

	user := &entity.AppleUser{}
	fields, values := user.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", strings.Join(fields, ","), user.TableName())

	err := db.QueryRow(ctx, query, &userID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return user, nil
}
