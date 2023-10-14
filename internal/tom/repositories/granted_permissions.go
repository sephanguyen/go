package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type GrantedPermissionsRepo struct {
}

func (g *GrantedPermissionsRepo) FindByUserIDAndPermissionName(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, permissionName pgtype.Text) (map[string][]*entities.GrantedPermission, error) {
	ctx, span := interceptors.StartSpan(ctx, "GrantedPermissionsRepo.FindByUserIDAndPermissionName")
	defer span.End()

	m := &entities.GrantedPermission{}
	fields := database.GetFieldNames(m)

	query := fmt.Sprintf(`SELECT %s FROM %s
		WHERE user_id = ANY($1)
		AND permission_name = $2`, strings.Join(fields, ","), m.TableName())

	rows, err := db.Query(ctx, query, userIDs, permissionName)
	if err != nil {
		return nil, fmt.Errorf("db.Query %w", err)
	}
	defer rows.Close()
	grantedLocationMap := make(map[string][]*entities.GrantedPermission)

	for rows.Next() {
		grantedPermission := &entities.GrantedPermission{}

		if err := rows.Scan(database.GetScanFields(grantedPermission, fields)...); err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}
		grantedLocationMap[grantedPermission.UserID.String] = append(grantedLocationMap[grantedPermission.UserID.String], grantedPermission)
	}
	return grantedLocationMap, nil
}
