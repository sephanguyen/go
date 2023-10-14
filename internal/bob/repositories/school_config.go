package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
)

type SchoolConfigRepo struct{}

func (rcv *SchoolConfigRepo) FindByID(ctx context.Context, db database.QueryExecer, id pgtype.Int4) (*entities.SchoolConfig, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolConfigRepo.FindByID")
	defer span.End()

	e := &entities.SchoolConfig{}
	fields, values := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE school_id = $1", strings.Join(fields, ","), e.TableName())

	err := db.QueryRow(ctx, query, &id).Scan(values...)
	if err != nil {
		return nil, errors.Wrap(err, "db.QueryRowEx")
	}

	return e, nil
}

func (rcv *SchoolConfigRepo) FindByIDs(ctx context.Context, db database.QueryExecer, ids pgtype.Int4Array) (map[pgtype.Int4]*entities.SchoolConfig, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolConfigRepo.FindByIDs")
	defer span.End()

	e := &entities.SchoolConfig{}
	fields, _ := e.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE school_id = ANY($1)", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	result := make(map[pgtype.Int4]*entities.SchoolConfig)
	for rows.Next() {
		e := &entities.SchoolConfig{}
		_, values := e.FieldMap()

		err = rows.Scan(values...)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		result[e.ID] = e
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}

	return result, nil
}
