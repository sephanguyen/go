package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// ConfigRepo to wrk with configs table
type ConfigRepo struct{}

// Retrieve return list of config specific by keys
func (c *ConfigRepo) Retrieve(ctx context.Context, db database.QueryExecer,
	country pgtype.Text, group pgtype.Text, keys pgtype.TextArray) ([]*entities_bob.Config, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.Retrieve")
	defer span.End()

	conf := &entities_bob.Config{}
	fieldNames, _ := conf.FieldMap()
	query := fmt.Sprintf("SELECT %s "+
		"FROM %s "+
		"WHERE config_key = ANY($1) AND config_group = $2 AND country = $3", strings.Join(fieldNames, ","), conf.TableName())
	rows, err := db.Query(ctx, query, &keys, &group, &country)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var configs []*entities_bob.Config
	for rows.Next() {
		cfg := &entities_bob.Config{}
		if err := rows.Scan(database.GetScanFields(cfg, fieldNames)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		configs = append(configs, cfg)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return configs, nil
}

func (c *ConfigRepo) Find(ctx context.Context, db database.QueryExecer,
	country pgtype.Text, group pgtype.Text) ([]*entities_bob.Config, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.Find")
	defer span.End()

	conf := &entities_bob.Config{}
	fieldNames, _ := conf.FieldMap()
	query := fmt.Sprintf("SELECT %s "+
		"FROM %s "+
		"WHERE config_group = $1 AND country = $2", strings.Join(fieldNames, ","), conf.TableName())
	rows, err := db.Query(ctx, query, &group, &country)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var configs []*entities_bob.Config
	for rows.Next() {
		cfg := &entities_bob.Config{}
		if err := rows.Scan(database.GetScanFields(cfg, fieldNames)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		configs = append(configs, cfg)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return configs, nil
}

func (c *ConfigRepo) RetrieveWithResourcePath(ctx context.Context, db database.QueryExecer,
	country pgtype.Text, group pgtype.Text, keys pgtype.TextArray, resourcePath pgtype.Text) ([]*entities_bob.Config, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.RetrieveWithResourcePath")
	defer span.End()

	conf := &entities_bob.Config{}
	fieldNames, _ := conf.FieldMap()
	query := fmt.Sprintf("SELECT %s "+
		"FROM %s "+
		"WHERE config_key = ANY($1) AND config_group = $2 AND country = $3 and resource_path = $4", strings.Join(fieldNames, ","), conf.TableName())
	rows, err := db.Query(ctx, query, &keys, &group, &country, &resourcePath)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	var configs []*entities_bob.Config
	for rows.Next() {
		cfg := &entities_bob.Config{}
		if err := rows.Scan(database.GetScanFields(cfg, fieldNames)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		configs = append(configs, cfg)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return configs, nil
}

func (c *ConfigRepo) Upsert(ctx context.Context, db database.Ext, cc []*entities_bob.Config) error {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.Upsert")
	defer span.End()

	queue := func(b *pgx.Batch, t *entities_bob.Config) {
		fieldNames := database.GetFieldNames(t)
		placeHolders := database.GeneratePlaceholders(len(fieldNames))

		query := fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s) ON CONFLICT ON CONSTRAINT config_pk DO UPDATE
		SET updated_at = now(), deleted_at = NULL, config_value = $4`, t.TableName(), strings.Join(fieldNames, ","), placeHolders)
		b.Queue(query, database.GetScanFields(t, fieldNames)...)
	}

	now := time.Now()
	b := &pgx.Batch{}

	for _, t := range cc {
		err := multierr.Combine(
			t.CreatedAt.Set(now),
			t.UpdatedAt.Set(now),
		)

		if err != nil {
			return fmt.Errorf("multierr.Err: %w", err)
		}

		queue(b, t)
	}
	batchResults := db.SendBatch(ctx, b)
	defer batchResults.Close()

	for i := 0; i < len(cc); i++ {
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
