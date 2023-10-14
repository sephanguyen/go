package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

type ExternalConfigRepo struct{}

func (e *ExternalConfigRepo) GetByKey(ctx context.Context, db database.QueryExecer, cKey string) (*domain.ExternalConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExternalConfigRepo.GetByKey")
	defer span.End()

	config := &domain.ExternalConfiguration{}
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

func (e *ExternalConfigRepo) GetByMultipleKeys(ctx context.Context, db database.QueryExecer, cKey []string) ([]*domain.ExternalConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExternalConfigRepo.GetByMultipleKeys")
	defer span.End()

	config := &domain.ExternalConfiguration{}
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

func (e *ExternalConfigRepo) SearchWithKey(ctx context.Context, db database.QueryExecer, payload domain.ExternalConfigSearchArgs) ([]*domain.ExternalConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExternalConfigRepo.SearchWithKey")
	defer span.End()

	hasKeyword := strings.TrimSpace(payload.Keyword) != ""
	keyFilter := ""
	if hasKeyword {
		keyFilter = fmt.Sprintf(` AND config_key ILIKE '%%%s%%'`, payload.Keyword)
	}

	config := &domain.ExternalConfiguration{}
	fields, _ := config.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE deleted_at IS NULL %s
		ORDER BY config_key DESC LIMIT $1 OFFSET $2`,
		strings.Join(fields, ","),
		config.TableName(),
		keyFilter,
	)
	rows, err := db.Query(ctx, query, database.Int8(payload.Limit), database.Int8(payload.Offset))
	if err != nil {
		return nil, err
	}
	return readConfigurations(rows)
}

func readConfigurations(rows pgx.Rows) ([]*domain.ExternalConfiguration, error) {
	var cfs []*domain.ExternalConfiguration
	config := &domain.ExternalConfiguration{}
	fields, _ := config.FieldMap()

	defer rows.Close()
	for rows.Next() {
		cf := new(domain.ExternalConfiguration)
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

func (e *ExternalConfigRepo) CreateMultipleConfigs(ctx context.Context, db database.QueryExecer, configs []*domain.ExternalConfiguration) error {
	ctx, span := interceptors.StartSpan(ctx, "ExternalConfigRepo.CreateMultipleConfigs")
	defer span.End()

	queue := func(b *pgx.Batch, t *domain.ExternalConfiguration) {
		query := fmt.Sprintf(`UPDATE %s SET config_value = $2, config_value_type = $3 WHERE config_key = $1`, t.TableName())
		b.Queue(query, t.ConfigKey, t.ConfigValue, t.ConfigValueType)
	}

	b := &pgx.Batch{}

	for _, t := range configs {
		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(configs); i++ {
		ct, err := batchResults.Exec()
		if err != nil {
			return fmt.Errorf("batchResults.Exec: %w", err)
		}
		if ct.RowsAffected() != 1 {
			return fmt.Errorf("configs not inserted")
		}
	}
	return nil
}

func (e *ExternalConfigRepo) GetByKeysAndLocations(ctx context.Context, db database.QueryExecer, configKeys, locationIDS []string) ([]*domain.LocationConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExternalConfigRepo.GetByKeysAndLocations")
	defer span.End()

	config := &domain.LocationConfiguration{}
	fields, _ := config.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE config_key = ANY($1) AND location_id = ANY($2)
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		config.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(configKeys), database.TextArray(locationIDS))
	if err != nil {
		return nil, err
	}

	var cfs []*domain.LocationConfiguration

	defer rows.Close()
	for rows.Next() {
		cf := new(domain.LocationConfiguration)
		if err := rows.Scan(database.GetScanFields(cf, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		cfs = append(cfs, cf)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return cfs, nil
}

func (e *ExternalConfigRepo) GetByKeysAndLocationsV2(ctx context.Context, db database.QueryExecer, configKeys, locationIDS []string) ([]*domain.LocationConfigurationV2, error) {
	ctx, span := interceptors.StartSpan(ctx, "ExternalConfigRepo.GetByKeysAndLocationsV2")
	defer span.End()

	config := &domain.LocationConfigurationV2{}
	fields, _ := config.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE config_key = ANY($1) AND location_id = ANY($2)
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		config.TableName(),
	)

	rows, err := db.Query(ctx, query, database.TextArray(configKeys), database.TextArray(locationIDS))
	if err != nil {
		return nil, err
	}

	var cfs []*domain.LocationConfigurationV2

	defer rows.Close()
	for rows.Next() {
		cf := new(domain.LocationConfigurationV2)
		if err := rows.Scan(database.GetScanFields(cf, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		cfs = append(cfs, cf)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return cfs, nil
}
