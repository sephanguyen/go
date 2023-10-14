package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/pkg/errors"
)

type DateTypeRepo struct{}

func (d *DateTypeRepo) GetDateTypeByID(ctx context.Context, db database.QueryExecer, id string) (*dto.DateType, error) {
	ctx, span := interceptors.StartSpan(ctx, "DateTypeRepo.GetDateTypeByID")
	defer span.End()

	dateType := &DateType{}
	fields, values := dateType.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE day_type_id = $1 AND deleted_at IS NULL `,
		strings.Join(fields, ","),
		dateType.TableName(),
	)
	if err := db.QueryRow(ctx, query, id).Scan(values...); err != nil {
		return nil, err
	}

	return dateType.ConvertToDTO(), nil
}

func (d *DateTypeRepo) GetDateTypeByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*dto.DateType, error) {
	ctx, span := interceptors.StartSpan(ctx, "DateTypeRepo.GetDateTypeByIDs")
	defer span.End()

	dateType := &DateType{}
	fields, _ := dateType.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s 
		WHERE day_type_id = ANY($1) AND deleted_at IS NULL `,
		strings.Join(fields, ","),
		dateType.TableName(),
	)

	rows, err := db.Query(ctx, query, ids)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	dateTypes := []*DateType{}
	for rows.Next() {
		dateType := &DateType{}
		if err := rows.Scan(database.GetScanFields(dateType, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		dateTypes = append(dateTypes, dateType)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	dateTypeList := make([]*dto.DateType, 0, len(dateTypes))
	for _, dateType := range dateTypes {
		dateTypeList = append(dateTypeList, dateType.ConvertToDTO())
	}

	return dateTypeList, nil
}
