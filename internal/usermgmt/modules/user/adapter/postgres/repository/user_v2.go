package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

// UserRepoV2 stores, should only be used by Shamir service
type UserRepoV2 struct{}

func (r *UserRepoV2) GetByAuthInfo(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.LegacyUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepoV2.GetByAuthInfo")
	defer span.End()

	u := &entity.LegacyUser{}
	fields, _ := u.FieldMap()

	groupFields := make([]string, 0, len(fields))
	for _, field := range fields {
		groupFields = append(groupFields, "users."+field)
	}

	joinStmt := ""

	if defaultOrganizationAuthValues == "" {
		joinStmt = fmt.Sprintf(
			`%s united_organization_auths`,
			(&entity.OrganizationAuth{}).TableName(),
		)
	} else {
		joinStmt = fmt.Sprintf(
			`
			(
				SELECT * FROM %s
    				UNION
				%s
			) AS united_organization_auths
			`,
			(&entity.OrganizationAuth{}).TableName(),
			defaultOrganizationAuthValues,
		)
	}

	// update to query from staff
	query := fmt.Sprintf(
		`
		SELECT 
			%[1]s
		FROM 
			%[2]s
		JOIN 
			%[3]s
		ON 
			split_part(%[2]s.resource_path, ':', 1) = CAST(united_organization_auths.organization_id AS text)
		WHERE 
			%[2]s.user_id = $1
			AND united_organization_auths.auth_project_id = $2 
			AND united_organization_auths.auth_tenant_id = $3
		GROUP BY 
			%[4]s,
			united_organization_auths.organization_id, 
			united_organization_auths.auth_project_id, 
			united_organization_auths.auth_tenant_id
		`,
		strings.Join(fields, ","),
		u.TableName(),
		joinStmt,
		strings.Join(groupFields, ","),
	)

	// ctxzap.Extract(ctx).Sugar().Debug("UserGroupRepoV2.Get", query)

	err := database.Select(ctx, db, query, &userID, &projectID, &tenantID).ScanOne(u)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return u, nil
}

func (r *UserRepoV2) GetByAuthInfoV2(ctx context.Context, db database.QueryExecer, defaultOrganizationAuthValues string, userID string, projectID string, tenantID string) (*entity.AuthUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepoV2.GetByAuthInfoV2")
	defer span.End()

	u := &entity.AuthUser{}
	fields, _ := u.FieldMap()

	groupFields := make([]string, 0, len(fields))
	for _, field := range fields {
		groupFields = append(groupFields, "users."+field)
	}

	joinStmt := ""

	if defaultOrganizationAuthValues == "" {
		joinStmt = fmt.Sprintf(
			`%s united_organization_auths`,
			(&entity.OrganizationAuth{}).TableName(),
		)
	} else {
		joinStmt = fmt.Sprintf(
			`
			(
				SELECT * FROM %s
    				UNION
				%s
			) AS united_organization_auths
			`,
			(&entity.OrganizationAuth{}).TableName(),
			defaultOrganizationAuthValues,
		)
	}

	// update to query from staff
	query := fmt.Sprintf(
		`
		SELECT 
			%[1]s
		FROM 
			%[2]s
		JOIN 
			%[3]s
		ON 
			split_part(%[2]s.resource_path, ':', 1) = CAST(united_organization_auths.organization_id AS text)
		WHERE 
			%[2]s.user_id = $1
			AND united_organization_auths.auth_project_id = $2 
			AND united_organization_auths.auth_tenant_id = $3
		GROUP BY 
			%[4]s,
			united_organization_auths.organization_id, 
			united_organization_auths.auth_project_id, 
			united_organization_auths.auth_tenant_id
		`,
		strings.Join(fields, ","),
		u.TableName(),
		joinStmt,
		strings.Join(groupFields, ","),
	)

	err := database.Select(ctx, db, query, &userID, &projectID, &tenantID).ScanOne(u)
	if err != nil {
		return nil, fmt.Errorf("database.Select: %w", err)
	}

	return u, nil
}

// GetByUsername to get auth user by username, because Shamir service is by pass rls, organizationID is mandatory
func (r *UserRepoV2) GetByUsername(ctx context.Context, db database.QueryExecer, username, organizationID string) (*entity.AuthUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepoV2.GetByUsername")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE username = $1 and resource_path = $2`

	user := new(entity.AuthUser)
	fieldNames, fieldValues := user.FieldMap()

	stmt = fmt.Sprintf(stmt, strings.Join(fieldNames, ","), user.TableName())

	zapLogger := ctxzap.Extract(ctx).Sugar()
	zapLogger.Debug(stmt)

	row := db.QueryRow(ctx, stmt, username, organizationID)

	if err := row.Scan(fieldValues...); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByEmail to get auth user by email, because Shamir service is by pass rls, organizationID is mandatory
func (r *UserRepoV2) GetByEmail(ctx context.Context, db database.QueryExecer, email, organizationID string) (*entity.AuthUser, error) {
	ctx, span := interceptors.StartSpan(ctx, "UserRepoV2.GetByEmail")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE email = $1 and resource_path = $2`

	user := new(entity.AuthUser)
	fieldNames, fieldValues := user.FieldMap()

	stmt = fmt.Sprintf(stmt, strings.Join(fieldNames, ","), user.TableName())

	zapLogger := ctxzap.Extract(ctx).Sugar()
	zapLogger.Debug(stmt)

	row := db.QueryRow(ctx, stmt, email, organizationID)

	if err := row.Scan(fieldValues...); err != nil {
		return nil, err
	}

	return user, nil
}
