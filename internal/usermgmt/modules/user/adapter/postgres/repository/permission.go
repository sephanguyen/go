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
	"go.uber.org/multierr"
)

type PermissionRepo struct{}

func (p *PermissionRepo) CreateBatch(ctx context.Context, db database.Ext, permissions []*entity.Permission) error {
	ctx, span := interceptors.StartSpan(ctx, "PermissionRepo.CreateBatch")
	defer span.End()

	queueFn := func(b *pgx.Batch, permission *entity.Permission) {
		fields, values := permission.FieldMap()
		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			permission.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := time.Now()

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	for _, permission := range permissions {
		if err := multierr.Combine(
			permission.CreatedAt.Set(now),
			permission.UpdatedAt.Set(now),
			permission.DeletedAt.Set(nil),
		); err != nil {
			return err
		}

		if permission.ResourcePath.Status == pgtype.Null {
			if err := permission.ResourcePath.Set(resourcePath); err != nil {
				return err
			}
		}

		queueFn(batch, permission)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for i := 0; i < len(permissions); i++ {
		cmdTag, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec %s", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("permission not inserted")
		}
	}

	return nil
}
