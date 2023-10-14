package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type LegacyUserGroupRepo struct{}

type LegacyUserGroup struct {
	entity.LegacyUserGroup

	// These attributes belong to postgres database context
	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func (*LegacyUserGroup) TableName() string {
	return "users_groups"
}

func (legacyUserGroup *LegacyUserGroup) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "group_id", "is_origin", "status", "created_at", "updated_at", "resource_path"}
	values = []interface{}{legacyUserGroup.UserID().Ptr(), legacyUserGroup.GroupID().Ptr(), legacyUserGroup.IsOrigin().Ptr(), legacyUserGroup.Status().Ptr(), legacyUserGroup.CreatedAt.Ptr(), legacyUserGroup.UpdatedAt.Ptr(), legacyUserGroup.OrganizationID().Ptr()}
	return
}

func (repo *LegacyUserGroupRepo) createMultiple(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error {
	ctx, span := interceptors.StartSpan(ctx, "LegacyUserGroupRepo.createMultiple")
	defer span.End()

	queueFn := func(b *pgx.Batch, legacyUserGroup *LegacyUserGroup) {
		fields, values := legacyUserGroup.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT pk__users_groups DO NOTHING",
			legacyUserGroup.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}
	now := field.NewTime(time.Now())

	for _, legacyUserGroup := range legacyUserGroups {
		repoLegacyUserGroup := &LegacyUserGroup{
			LegacyUserGroup: legacyUserGroup,
			UpdatedAt:       now,
			CreatedAt:       now,
			DeletedAt:       field.NewNullTime(),
		}

		queueFn(batch, repoLegacyUserGroup)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(legacyUserGroups); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}
