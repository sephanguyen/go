package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type UserDeviceTokenRepo struct{}

func (repo *UserDeviceTokenRepo) UpsertUserDeviceToken(ctx context.Context, db database.QueryExecer, u *entities.UserDeviceToken) error {
	ctx, span := interceptors.StartSpan(ctx, "UserDeviceTokenRepo.Upsert")
	defer span.End()

	now := time.Now()
	err := multierr.Combine(
		u.UpdatedAt.Set(now),
		u.CreatedAt.Set(now),
	)

	if err != nil {
		return fmt.Errorf("multierr.Combine: %w", err)
	}

	fields := []string{"user_id", "device_token", "allow_notification", "created_at", "updated_at"}
	placeHolders := "$1, $2, $3, $4, $5"

	upsertStm := ""
	if u.DeviceToken.Status == pgtype.Present {
		upsertStm += ", device_token = $2"
	}
	if u.AllowNotification.Status == pgtype.Present {
		upsertStm += ", allow_notification = $3"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT user_id_un DO UPDATE SET updated_at = $5"+upsertStm, u.TableName(), strings.Join(fields, ", "), placeHolders)
	args := database.GetScanFields(u, fields)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return errors.Wrap(err, "r.DB.ExecEx")
	}
	return nil
}

func (repo *UserDeviceTokenRepo) FindByUserIDs(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) (entities.UserDeviceTokens, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserDeviceTokenRepo.FindByUserIDs")
	defer span.End()
	e := &entities.UserDeviceToken{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf(`
	SELECT %s
	FROM user_device_tokens udt
	WHERE udt.user_id = ANY($1)
	`, strings.Join(fields, ","))
	ents := entities.UserDeviceTokens{}
	err := database.Select(ctx, db, query, userIDs).ScanAll(&ents)
	if err != nil {
		return nil, err
	}
	return ents, nil
}
