package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type CityRepo struct{}

func (r *CityRepo) Get(ctx context.Context, db database.QueryExecer, country, name string) ([]*entities.City, error) {
	var p entities.Citites

	e := &entities.City{}
	fieldNames := database.GetFieldNames(e)
	stmt := "SELECT %s FROM %s WHERE country = $1 AND NAME LIKE ('%%' || $2 || '%%')"
	query := fmt.Sprintf(stmt, strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, query, country, name).ScanAll(&p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
