package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type UserDeviceTokenRepo struct {
}

func (r *UserDeviceTokenRepo) Upsert(ctx context.Context, db database.QueryExecer, u *entities.UserDeviceToken) error {
	ctx, span := interceptors.StartSpan(ctx, "UserDeviceTokenRepo.Upsert")
	defer span.End()

	now := time.Now()
	u.UpdatedAt.Set(now)
	u.CreatedAt.Set(now)

	fields := []string{"user_id", "token", "allow_notification", "created_at", "updated_at", "user_name"}
	placeHolders := generateInsertPlaceholders(len(fields))

	upsertStm := ""
	if u.Token.Status == pgtype.Present {
		upsertStm += ", token = $2"
	}

	if u.AllowNotification.Status == pgtype.Present {
		upsertStm += ", allow_notification = $3"
	}

	if u.UserName.Status == pgtype.Present {
		upsertStm += ", user_name = $6"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT user_id_un DO UPDATE SET updated_at = $5"+upsertStm, u.TableName(), strings.Join(fields, ", "), placeHolders)
	args := database.GetScanFields(u, fields)

	if _, err := db.Exec(ctx, query, args...); err != nil {
		return errors.Wrap(err, "r.DB.ExecEx")
	}
	return nil
}

func (r *UserDeviceTokenRepo) Find(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserDeviceTokenRepo.FindDeviceTokenByUser")
	defer span.End()

	c := &entities.UserDeviceToken{}
	selectStmt := fmt.Sprintf("SELECT token FROM %s WHERE user_id = ANY($1) AND allow_notification IS TRUE", c.TableName())

	rows, err := db.Query(ctx, selectStmt, &userIDs)
	if err != nil {
		return nil, errors.Wrap(err, "r.DB.QueryEx")
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var deviceToken pgtype.Text
		if err := rows.Scan(&deviceToken); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}

		result = append(result, deviceToken.String)
	}

	return result, nil
}

func (r *UserDeviceTokenRepo) FindByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (c *entities.UserDeviceToken, err error) {
	ctx, span := interceptors.StartSpan(ctx, "ConversationRepo.FindByID")
	defer span.End()

	c = new(entities.UserDeviceToken)
	fields := database.GetFieldNames(c)
	selectStmt := fmt.Sprintf("SELECT %s FROM %s WHERE user_id = $1", strings.Join(fields, ","), c.TableName())

	row := db.QueryRow(ctx, selectStmt, &userID)

	if err := row.Scan(database.GetScanFields(c, fields)...); err != nil {
		return nil, errors.Wrapf(err, "row.Scan: userID: %q", userID.String)
	}

	return
}

func (r *UserDeviceTokenRepo) BulkUpdateResourcePath(ctx context.Context, db database.QueryExecer, userIDs []string, resourcePath string) error {
	stmt := `
update user_device_tokens u set resource_path = $1
where u.user_id = ANY($2)
and (u.resource_path is null or length(u.resource_path)=0)`
	_, err := db.Exec(ctx, stmt, database.Text(resourcePath), database.TextArray(userIDs))
	if err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}
