package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ConfigRepo struct{}

func (c *ConfigRepo) GetByKey(ctx context.Context, db database.QueryExecer, cKey string) (*domain.InternalConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.GetByKey")
	defer span.End()

	config := &domain.InternalConfiguration{}
	fields, values := config.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE config_key = $1
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		config.TableName(),
	)

	if err := db.QueryRow(ctx, query, cKey).Scan(values...); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *ConfigRepo) GetByMultipleKeys(ctx context.Context, db database.QueryExecer, cKey []string) ([]*domain.InternalConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.GetByMultipleKeys")
	defer span.End()

	config := &domain.InternalConfiguration{}
	fields, _ := config.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE config_key = ANY($1)
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		config.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(cKey))
	if err != nil {
		return nil, err
	}
	return readConfigurations(rows)
}

func (c *ConfigRepo) SearchWithKey(ctx context.Context, db database.QueryExecer, payload domain.ConfigSearchArgs) ([]*domain.InternalConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.SearchWithKey")
	defer span.End()

	hasKeyword := strings.TrimSpace(payload.Keyword) != ""
	keyFilter := ""
	if hasKeyword {
		keyFilter = fmt.Sprintf(` AND config_key ILIKE '%%%s%%'`, payload.Keyword)
	}

	config := &domain.InternalConfiguration{}
	fields, _ := config.FieldMap()

	exConfig := &domain.ExternalConfiguration{}
	exFields, _ := exConfig.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s WHERE deleted_at IS NULL %s UNION SELECT %s FROM %s WHERE deleted_at IS NULL %s ORDER BY config_key DESC LIMIT $1 OFFSET $2`,
		strings.Join(fields, ","),
		config.TableName(),
		keyFilter,
		strings.Join(exFields, ","),
		exConfig.TableName(),
		keyFilter,
	)

	rows, err := db.Query(ctx, query, database.Int8(payload.Limit), database.Int8(payload.Offset))
	if err != nil {
		return nil, err
	}
	return readConfigurations(rows)
}

func readConfigurations(rows pgx.Rows) ([]*domain.InternalConfiguration, error) {
	var cfs []*domain.InternalConfiguration
	config := &domain.InternalConfiguration{}
	fields, _ := config.FieldMap()

	defer rows.Close()
	for rows.Next() {
		cf := new(domain.InternalConfiguration)
		if err := rows.Scan(database.GetScanFields(cf, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		cfs = append(cfs, cf)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return cfs, nil
}
