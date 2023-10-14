package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/crypt"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

const ResourcePathFieldName = "resource_path"

// OrganizationRepo provides method to work with parent entity
type OrganizationRepo struct {
	defaultOrganizationAuthValues string
}

func (r *OrganizationRepo) WithDefaultValue(env string) *OrganizationRepo {
	r.defaultOrganizationAuthValues = r.DefaultOrganizationAuthValues(env)
	return r
}

func (r *OrganizationRepo) DefaultOrganizationAuthValues(env string) string {
	var stmt string
	switch env {
	case "local":
		stmt =
			`
			VALUES 
				(-2147483648, 'dev-manabie-online', ''),
				(-2147483648, 'dev-manabie-online', 'manabie-0nl6t'),
				(-2147483647, 'manabie-test', ''),
				(-2147483647, 'dev-manabie-online', 'jprep-eznr7'),
				(-2147483646, 'dev-manabie-online', ''),
				(-2147483646, 'dev-manabie-online', 'synersia-24rue'),
				(-2147483645, 'dev-manabie-online', ''),
				(-2147483645, 'dev-manabie-online', 'renseikai-yu9y7'),
				(-2147483644, 'dev-manabie-online', ''),
				(-2147483644, 'dev-manabie-online', 'end-to-end-dopvo'),
				(-2147483643, 'dev-manabie-online', ''),
				(-2147483643, 'dev-manabie-online', 'ga-school-jhe90'),
				(-2147483642, 'dev-manabie-online', ''),
				(-2147483642, 'dev-manabie-online', 'kec-school-ovmgv'),
				(-2147483641, 'dev-manabie-online', ''),
				(-2147483641, 'dev-manabie-online', 'aic-school-5qbbu'),
				(-2147483640, 'dev-manabie-online', ''),
				(-2147483640, 'dev-manabie-online', 'nsg-school-6osx0'),
				(-2147483630, 'dev-manabie-online', ''),
				(-2147483630, 'dev-manabie-online', 'withus-managara-base-0wf23'),
				(-2147483629, 'dev-manabie-online', ''),
				(-2147483629, 'dev-manabie-online', 'withus-managara-hs-t5fuk'),
				(1, 'dev-manabie-online', 'integration-test-1-909wx')
			`
	case "stag":
		stmt =
			`
			VALUES 
				(-2147483648, 'staging-manabie-online', ''),
				(-2147483648, 'staging-manabie-online', 'manabie-p7muf'),
				(-2147483647, 'staging-manabie-online', ''),
				(-2147483647, 'manabie-test', ''),
				(-2147483646, 'staging-manabie-online', ''),
				(-2147483646, 'staging-manabie-online', 'synersia-m3hr5'),
				(-2147483645, 'staging-manabie-online', ''),
				(-2147483645, 'staging-manabie-online', 'renseikai-5ayyd'),
				(-2147483644, 'staging-manabie-online', ''),
				(-2147483644, 'staging-manabie-online', 'end-to-end-school-5xn27'),
				(-2147483643, 'staging-manabie-online', ''),
				(-2147483643, 'staging-manabie-online', 'ga-school-rtaas'),
				(-2147483642, 'staging-manabie-online', ''),
				(-2147483642, 'staging-manabie-online', 'kec-school-q148v'),
				(-2147483641, 'staging-manabie-online', ''),
				(-2147483641, 'staging-manabie-online', 'aic-school-xhy07'),
				(-2147483640, 'staging-manabie-online', ''),
				(-2147483640, 'staging-manabie-online', 'nsg-school-5wkab'),
				(16091, 'staging-manabie-online', ''),
				(16091, 'staging-manabie-online', 'manabie2-edzop')
			`
	case "uat":
		stmt =
			`
			VALUES 
				(-2147483648, 'uat-manabie', ''),
				(-2147483648, 'uat-manabie', 'manabie-9h0ng'),
				(-2147483647, 'uat-manabie', ''),
				(-2147483647, 'manabie-test', ''),
				(-2147483646, 'uat-manabie', ''),
				(-2147483646, 'uat-manabie', 'synersia-oodzl'),
				(-2147483645, 'uat-manabie', ''),
				(-2147483645, 'uat-manabie', 'renseikai-zxx25'),
				(-2147483644, 'uat-manabie', ''),
				(-2147483644, 'uat-manabie', 'end-to-end-school-5mqoc'),
				(-2147483643, 'uat-manabie', ''),
				(-2147483643, 'uat-manabie', 'ga-school-q3mvu'),
				(-2147483642, 'uat-manabie', ''),
				(-2147483642, 'uat-manabie', 'kec-school-8qe69'),
				(-2147483641, 'uat-manabie', ''),
				(-2147483641, 'uat-manabie', 'aic-school-fj80n'),
				(-2147483640, 'uat-manabie', ''),
				(-2147483640, 'uat-manabie', 'nsg-school-yevm8'),
				(16091, 'uat-manabie', ''),
				(16091, 'uat-manabie', 'manabie2-44xpg')
			`
	case "prod":
		stmt =
			`
			VALUES 
				(-2147483648, 'production-manabie-vn', ''),
				(-2147483648, 'production-manabie-vn', 'prod-manabie-bj1ok'),
				(-2147483647, 'jprep', ''),
				(-2147483646, 'synersia', ''),
				(-2147483646, 'synersia', 'prod-synersia-trscc'),
				(-2147483645, 'production-renseikai', ''),
				(-2147483645, 'production-renseikai', 'prod-renseikai-8xr29'),
				(-2147483643, 'production-ga', ''),
				(-2147483643, 'production-ga', 'prod-ga-uq2rq'),
				(-2147483642, 'production-kec', ''),
				(-2147483642, 'production-kec', 'prod-kec-58ww0'),
				(-2147483641, 'production-aic', ''),
				(-2147483641, 'production-aic', 'prod-aic-u3d1m'),
				(-2147483640, 'production-nsg', ''),
				(-2147483640, 'production-nsg', 'prod-nsg-flbh7')
			`
	}
	return stmt
}

func (r *OrganizationRepo) GetTenantIDByOrgID(ctx context.Context, db database.QueryExecer, orgID string) (string, error) {
	unionDefaultValues := ``
	if r.defaultOrganizationAuthValues != "" {
		unionDefaultValues = fmt.Sprintf(`UNION %s`, r.defaultOrganizationAuthValues)
	}
	stmt := fmt.Sprintf(
		`
		WITH cte (organization_id, auth_project_id, auth_tenant_id) AS (
			SELECT
				organization_id, auth_project_id, auth_tenant_id
			FROM
				%s
			WHERE
				organization_auths.auth_tenant_id != ''
			%s
		)
		SELECT
			cte.auth_tenant_id
		FROM
			cte
		WHERE
			cte.auth_tenant_id != '' AND cte.organization_id::text = $1
		`,
		(&entity.OrganizationAuth{}).TableName(),
		unionDefaultValues,
	)

	var tenantID string

	err := db.QueryRow(ctx, stmt, &orgID).Scan(&tenantID)

	switch err {
	case nil:
		return tenantID, nil
	case pgx.ErrNoRows:
		return "", nil
	default:
		return "", fmt.Errorf("row.Scan: %w", err)
	}
}

func (r *OrganizationRepo) GetAll(ctx context.Context, db database.QueryExecer, limit int) ([]*entity.OrganizationAuth, error) {
	unionDefaultValues := ``
	if r.defaultOrganizationAuthValues != "" {
		unionDefaultValues = fmt.Sprintf(`UNION %s`, r.defaultOrganizationAuthValues)
	}

	fieldNames, _ := (&entity.OrganizationAuth{}).FieldMap()

	fieldNamesWithPrefix := make([]string, 0, len(fieldNames))
	for _, fieldName := range fieldNames {
		fieldNamesWithPrefix = append(fieldNamesWithPrefix, "cte."+fieldName)
	}

	stmt := fmt.Sprintf(
		`
		WITH cte (organization_id, auth_project_id, auth_tenant_id) AS (
			SELECT
				%s
			FROM
				%s
			%s
			ORDER BY
				organization_id
			LIMIT
				$1
		)
		SELECT
			%s
		FROM
			cte
		LIMIT
			$1
		`,
		strings.Join(fieldNames, ", "),
		(&entity.OrganizationAuth{}).TableName(),
		unionDefaultValues,
		strings.Join(fieldNamesWithPrefix, ", "),
	)

	rows, err := db.Query(ctx, stmt, limit)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query()")
	}
	if rows.Err() != nil {
		return nil, errors.Wrap(err, "rows.Err()")
	}

	defer rows.Close()

	organizationAuths := make([]*entity.OrganizationAuth, 0, 64)

	for rows.Next() {
		organizationAuth := &entity.OrganizationAuth{}
		_, fieldValues := organizationAuth.FieldMap()

		if err := rows.Scan(fieldValues...); err != nil {
			return nil, errors.Wrap(err, "db.Query()")
		}

		organizationAuths = append(organizationAuths, organizationAuth)
	}

	return organizationAuths, nil
}

func (r *OrganizationRepo) GetByTenantID(ctx context.Context, db database.QueryExecer, tenantID string) (*entity.Organization, error) {
	stmt := `SELECT %s FROM %s WHERE tenant_id = $1`

	org := new(entity.Organization)
	fieldNames, fieldValues := org.FieldMap()

	stmt = fmt.Sprintf(stmt, strings.Join(fieldNames, ","), org.TableName())

	zapLogger := ctxzap.Extract(ctx).Sugar()
	zapLogger.Debug(stmt)

	row := db.QueryRow(ctx, stmt, &tenantID)

	if err := row.Scan(fieldValues...); err != nil {
		return nil, errors.Wrap(err, "Scan")
	}

	return org, nil
}

func (r *OrganizationRepo) GetByDomainName(ctx context.Context, db database.QueryExecer, domainName string) (*entity.Organization, error) {
	stmt := `SELECT %s FROM %s WHERE domain_name = $1`

	org := new(entity.Organization)
	fieldNames, fieldValues := org.FieldMap()

	stmt = fmt.Sprintf(stmt, strings.Join(fieldNames, ","), org.TableName())

	zapLogger := ctxzap.Extract(ctx).Sugar()
	zapLogger.Debug(stmt)

	row := db.QueryRow(ctx, stmt, &domainName)

	if err := row.Scan(fieldValues...); err != nil {
		return nil, err
	}

	return org, nil
}

func (r *OrganizationRepo) Find(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Organization, error) {
	ctx, span := interceptors.StartSpan(ctx, "OrganizationRepo.Find")
	defer span.End()

	organization := &entity.Organization{}
	fields := database.GetFieldNames(organization)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE organization_id = $1", strings.Join(fields, ","), organization.TableName())
	row := db.QueryRow(ctx, query, &id)
	if err := row.Scan(database.GetScanFields(organization, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return organization, nil
}

type TenantConfigRepo struct {
	database.QueryExecer

	ConfigAESKey string
	ConfigAESIv  string

	OrganizationRepo interface {
		GetByTenantID(ctx context.Context, db database.QueryExecer, tenantID string) (*entity.Organization, error)
	}
}

func (r *TenantConfigRepo) GetTenantConfig(ctx context.Context, tenantID string) (*gcp.TenantConfig, error) {
	if r == nil {
		return nil, errors.New("TenantConfigRepo is nil")
	}

	if r.QueryExecer == nil {
		return nil, errors.New("QueryExecer is nil")
	}

	if r.OrganizationRepo == nil {
		return nil, errors.New("OrganizationRepo is nil")
	}

	organization, err := r.OrganizationRepo.GetByTenantID(ctx, r.QueryExecer, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "GetByTenantID()")
	}

	key, err := crypt.DecodeBase64(r.ConfigAESKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode key with DecodeBase64()")
	}

	iv, err := crypt.DecodeBase64(r.ConfigAESIv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode iv with DecodeBase64()")
	}

	scryptSignerKey, err := crypt.AESDecryptBase64(organization.ScryptSignerKey.String, key, iv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode scryptSignerKey with AESDecrypt()")
	}
	scryptSignerKeyBytes, err := crypt.DecodeBase64(string(scryptSignerKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decodeBas64 scryptSignerKeyBytes")
	}
	scryptSaltSeparator, err := crypt.AESDecryptBase64(organization.ScryptSaltSeparator.String, key, iv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode scryptSaltSeparator with AESDecrypt()")
	}
	scryptSaltSeparatorBytes, err := crypt.DecodeBase64(string(scryptSaltSeparator))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decodeBas64 scryptSaltSeparator")
	}
	scryptRounds, err := crypt.AESDecryptBase64(organization.ScryptRounds.String, key, iv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode scryptRounds with AESDecrypt()")
	}
	scryptRoundsInt, err := strconv.ParseInt(string(scryptRounds), 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse scryptRounds with ParseInt()")
	}
	scryptMemoryCost, err := crypt.AESDecryptBase64(organization.ScryptMemoryCost.String, key, iv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode scryptMemoryCost with AESDecryptBase64()")
	}
	scryptMemoryCostInt, err := strconv.ParseInt(string(scryptMemoryCost), 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse scryptMemoryCost with ParseInt()")
	}

	tenantConfig := &gcp.TenantConfig{
		HashConfig: &gcp.HashConfig{
			HashAlgorithm: "SCRYPT",
			HashSignerKey: gcp.Base64EncodedStr{
				Value:        string(scryptSignerKey),
				DecodedBytes: scryptSignerKeyBytes,
			},
			HashSaltSeparator: gcp.Base64EncodedStr{
				Value:        string(scryptSaltSeparator),
				DecodedBytes: scryptSaltSeparatorBytes,
			},
			HashRounds:     int(scryptRoundsInt),
			HashMemoryCost: int(scryptMemoryCostInt),
		},
	}

	return tenantConfig, nil
}
