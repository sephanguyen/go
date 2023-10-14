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

type DomainUserAccessPathRepo struct{}

type UserAccessPath struct {
	UserIDAttr         field.String
	LocationIDAttr     field.String
	OrganizationIDAttr field.String

	// These attributes belong to postgres database context
	UpdatedAt field.Time
	CreatedAt field.Time
	DeletedAt field.Time
}

func (userAccessPath *UserAccessPath) UserID() field.String {
	return userAccessPath.UserIDAttr
}

func (userAccessPath *UserAccessPath) LocationID() field.String {
	return userAccessPath.LocationIDAttr
}

func (userAccessPath *UserAccessPath) OrganizationID() field.String {
	return userAccessPath.OrganizationIDAttr
}

func newDomainUserAccessPath(uap entity.DomainUserAccessPath) *UserAccessPath {
	now := field.NewTime(time.Now())
	return &UserAccessPath{
		UserIDAttr:         uap.UserID(),
		LocationIDAttr:     uap.LocationID(),
		OrganizationIDAttr: uap.OrganizationID(),
		UpdatedAt:          now,
		CreatedAt:          now,
		DeletedAt:          field.NewNullTime(),
	}
}

func (userAccessPath *UserAccessPath) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id", "location_id", "updated_at", "created_at", "deleted_at", "resource_path",
		}, []interface{}{
			&userAccessPath.UserIDAttr, &userAccessPath.LocationIDAttr, &userAccessPath.UpdatedAt, &userAccessPath.CreatedAt, &userAccessPath.DeletedAt, &userAccessPath.OrganizationIDAttr,
		}
}

func (userAccessPath *UserAccessPath) TableName() string {
	return "user_access_paths"
}

func (repo *DomainUserAccessPathRepo) GetByUserID(ctx context.Context, db database.QueryExecer, userID field.String) (entity.DomainUserAccessPaths, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserAccessPathRepo.GetByUserID")
	defer span.End()

	tableName := new(UserAccessPath).TableName()
	fieldNames, _ := new(UserAccessPath).FieldMap()

	statement :=
		`
		SELECT 
			%s 
		FROM 
		    %s 
		WHERE 
		    user_id = $1 AND 
		    deleted_at IS NULL
		`

	statement = fmt.Sprintf(
		statement,
		strings.Join(fieldNames, ","),
		tableName,
	)

	rows, err := db.Query(ctx, statement, &userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userAccessPaths := make(entity.DomainUserAccessPaths, 0, 100)
	for rows.Next() {
		userAccessPath := newDomainUserAccessPath(entity.DefaultUserAccessPath{})
		_, fieldValues := userAccessPath.FieldMap()

		if err := rows.Scan(fieldValues...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		userAccessPaths = append(userAccessPaths, userAccessPath)
	}

	return userAccessPaths, nil
}

func (repo *DomainUserAccessPathRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, userAccessPaths ...entity.DomainUserAccessPath) error {
	ctx, span := interceptors.StartSpan(ctx, "UserGroupRepo.upsert")
	defer span.End()

	queueFn := func(b *pgx.Batch, userAccessPath *UserAccessPath) {
		fields, values := userAccessPath.FieldMap()

		placeHolders := database.GeneratePlaceholders(len(fields))
		stmt := fmt.Sprintf(`
		INSERT INTO %s (%s) VALUES (%s)
		ON CONFLICT ON CONSTRAINT user_access_paths_pk 
		DO UPDATE SET updated_at = now(), deleted_at = NULL`,
			userAccessPath.TableName(),
			strings.Join(fields, ","),
			placeHolders,
		)

		b.Queue(stmt, values...)
	}

	batch := &pgx.Batch{}

	for _, userAccessPath := range userAccessPaths {
		repoUserAccessPath := newDomainUserAccessPath(userAccessPath)

		queueFn(batch, repoUserAccessPath)
	}

	batchResults := db.SendBatch(ctx, batch)
	defer func() {
		_ = batchResults.Close()
	}()

	for i := 0; i < len(userAccessPaths); i++ {
		_, err := batchResults.Exec()
		if err != nil {
			return errors.Wrap(err, "batchResults.Exec")
		}
	}

	return nil
}

func (repo *DomainUserAccessPathRepo) SoftDeleteByUserIDAndLocationIDs(ctx context.Context, db database.QueryExecer, userID, organizationID string, locationIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserAccessPathRepo.SoftDeleteByUserIDAndLocationIDs")
	defer span.End()

	sql := `UPDATE user_access_paths SET deleted_at = now(), updated_at = now() 
                         WHERE user_id = $1 
                           AND location_id = ANY($2) 
                           AND deleted_at IS NULL 
                           AND resource_path = $3`
	_, err := db.Exec(ctx, sql, database.Text(userID), database.TextArray(locationIDs), database.Text(organizationID))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (repo *DomainUserAccessPathRepo) SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) error {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserAccessPathRepo.SoftDeleteByUserIDs")
	defer span.End()

	uap := UserAccessPath{}

	sql := fmt.Sprintf(`UPDATE %s SET deleted_at = now() WHERE user_id = ANY($1) AND deleted_at IS NULL`, uap.TableName())
	_, err := db.Exec(ctx, sql, database.TextArray(userIDs))
	if err != nil {
		return fmt.Errorf("err db.Exec: %w", err)
	}

	return nil
}

func (repo *DomainUserAccessPathRepo) GetByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.DomainUserAccessPaths, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainUserAccessPathRepo.GetByUserIDs")
	defer span.End()

	uap := UserAccessPath{}
	fields, _ := uap.FieldMap()

	query := fmt.Sprintf(
		`
		SELECT %s
		FROM %s
		WHERE
			user_id = ANY($1) AND
			deleted_at IS NULL
		`,
		strings.Join(fields, ","),
		uap.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(userIDs))
	if err != nil {
		return nil, fmt.Errorf("db.Query: %v", err)
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %v", err)
	}

	userAccessPaths := make(entity.DomainUserAccessPaths, 0, len(userIDs))
	for rows.Next() {
		userAccessPath := newDomainUserAccessPath(&entity.DefaultUserAccessPath{})
		_, fields := userAccessPath.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %v", err)
		}
		userAccessPaths = append(userAccessPaths, userAccessPath)
	}

	return userAccessPaths, nil
}
