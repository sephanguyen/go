package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type ConfigRepo struct{}

func (c *ConfigRepo) GetConfigWithResourcePath(ctx context.Context, db database.QueryExecer, country domain.Country, group string, keys []string, resourcePath string) ([]*domain.Config, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.GetConfigWithResourcePath")
	defer span.End()

	countryString := string(country)
	config := &Config{}
	fields, values := config.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
		WHERE deleted_at IS NULL
		AND config_key = ANY($1) AND config_group = $2
		AND country = $3 AND resource_path = $4`,
		strings.Join(fields, ","),
		config.TableName(),
	)

	rows, err := db.Query(ctx, query, &keys, &group, &countryString, &resourcePath)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var configs []*domain.Config
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		configs = append(configs, config.ToConfigDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return configs, nil
}
