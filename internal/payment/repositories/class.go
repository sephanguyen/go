package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type ClassRepo struct {
}

func (r *ClassRepo) GetMapClassWithLocationByClassIDs(ctx context.Context, db database.QueryExecer, classIDs []string) (mapClass map[string]entities.Class, err error) {
	mapClass = make(map[string]entities.Class)
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			class_id = ANY($1) AND deleted_at IS NULL
		`
	fieldNames, _ := (&entities.Class{}).FieldMap()
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fieldNames, ","),
		(&entities.Class{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt, classIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		tmpClass := &entities.Class{}
		_, fieldValues := tmpClass.FieldMap()
		err = rows.Scan(fieldValues...)
		if err != nil {
			err = fmt.Errorf("row.Scan: %w", err)
			return
		}
		mapClass[tmpClass.ClassID.String] = *tmpClass
	}
	return
}
