package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

// SchoolAdminRepo works with school_admin_id
type SchoolAdminRepo struct{}

// Get returns school admin data if found
func (r *SchoolAdminRepo) Get(ctx context.Context, db database.QueryExecer, schoolAdminID pgtype.Text) (*entities.SchoolAdmin, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdminRepo.Get")
	defer span.End()

	e := &entities.SchoolAdmin{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE school_admin_id = $1", strings.Join(fields, ","), e.TableName())
	if err := database.Select(ctx, db, query, &schoolAdminID).ScanOne(e); err != nil {
		return nil, err
	}

	return e, nil
}

func (r *SchoolAdminRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, schoolAdmins []*entities.SchoolAdmin) error {
	ctx, span := interceptors.StartSpan(ctx, "SchoolAdmin.CreateMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, u *entities.SchoolAdmin) {
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

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	for _, u := range schoolAdmins {
		_ = u.UpdatedAt.Set(now)
		_ = u.CreatedAt.Set(now)
		if u.ResourcePath.Status != pgtype.Present {
			_ = u.ResourcePath.Set(resourcePath)
		}
		queueFn(b, u)
	}

	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(schoolAdmins); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}

		if ct.RowsAffected() != 1 {
			return fmt.Errorf("schoolAdmin not inserted")
		}
	}

	return nil
}
