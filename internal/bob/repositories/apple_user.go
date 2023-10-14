package repositories

import (
	"context"
	"fmt"
	"strings"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/timeutil"

	"github.com/jackc/pgtype"
)

type AppleUserRepo struct{}

func (r *AppleUserRepo) Create(ctx context.Context, db database.QueryExecer, a *entities_bob.AppleUser) error {
	ctx, span := interceptors.StartSpan(ctx, "AppleUserRepo.Create")
	defer span.End()

	now := timeutil.Now()
	_ = a.CreatedAt.Set(now)
	_ = a.UpdatedAt.Set(now)

	if _, err := database.Insert(ctx, a, db.Exec); err != nil {
		return fmt.Errorf("Insert: %w", err)
	}

	return nil
}

func (r *AppleUserRepo) Get(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entities_bob.AppleUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "AppleUserRepo.Get")
	defer span.End()

	e := &entities_bob.AppleUser{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &userID).Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("db.QueryRow: %w", err)
	}

	return e, nil
}
