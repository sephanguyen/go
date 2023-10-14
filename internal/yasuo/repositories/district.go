package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type DistrictRepo struct{}

// func (r *DistrictRepo) Get(ctx context.Context, db orm.DB, cityID int32, name string) ([]*entities.District, error) {
// 	var p []*entities.District

// 	err := db.ModelContext(ctx, &p).Where("city_id = ?", cityID).Where("name LIKE ?", "%"+name+"%").Select()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return p, nil
// }

func (r *DistrictRepo) Get(ctx context.Context, db database.QueryExecer, cityID int32, name string) ([]*entities.District, error) {
	var p entities.Districts

	e := &entities.District{}
	fieldNames := database.GetFieldNames(e)
	stmt := "SELECT %s FROM %s WHERE city_id = $1 AND name LIKE ('%%' || $2 || '%%')"
	query := fmt.Sprintf(stmt, strings.Join(fieldNames, ", "), e.TableName())
	err := database.Select(ctx, db, query, cityID, name).ScanAll(&p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
