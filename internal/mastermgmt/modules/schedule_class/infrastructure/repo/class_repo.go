package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type ClassRepo struct{}

func (c *ClassRepo) GetMapClassByIDs(ctx context.Context, db database.Ext, ids []string) (map[string]*Class, error) {
	ctx, span := interceptors.StartSpan(ctx, "ClassRepo.GetMapClassByIDs")
	defer span.End()
	classDTO := &Class{}

	fields := database.GetFieldNames(classDTO)
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE class_id = ANY($1)
			AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		classDTO.TableName(),
	)

	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mapClass := make(map[string]*Class)
	for rows.Next() {
		class := &Class{}
		_, value := class.FieldMap()
		if err = rows.Scan(value...); err != nil {
			return nil, err
		}
		mapClass[class.ClassID.String] = class
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return mapClass, nil
}
