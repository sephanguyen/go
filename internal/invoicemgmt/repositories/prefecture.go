package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"github.com/jackc/pgx/v4"
)

type PrefectureRepo struct {
}

func (repo PrefectureRepo) FindByPrefectureCode(ctx context.Context, db database.QueryExecer, prefectureCode string) (*entities.Prefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "PrefectureRepo.FindByPrefectureCode")
	defer span.End()

	prefecture := &entities.Prefecture{}
	fields, _ := prefecture.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE prefecture_code = $1 AND deleted_at IS NULL", strings.Join(fields, ","), prefecture.TableName())

	err := database.Select(ctx, db, query, prefectureCode).ScanOne(prefecture)

	switch err {
	case nil:
		return prefecture, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByID PrefectureRepo: %w", err)
	}
}

func (repo *PrefectureRepo) FindAll(ctx context.Context, db database.QueryExecer) ([]*entities.Prefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "PrefectureRepo.FindAll")
	defer span.End()

	e := &entities.Prefecture{}
	fields, _ := e.FieldMap()

	stmt := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	prefectures := []*entities.Prefecture{}
	defer rows.Close()
	for rows.Next() {
		prefecture := new(entities.Prefecture)
		database.AllNullEntity(prefecture)

		_, fieldValues := prefecture.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		prefectures = append(prefectures, prefecture)
	}

	return prefectures, nil
}

func (repo PrefectureRepo) FindByPrefectureID(ctx context.Context, db database.QueryExecer, prefectureID string) (*entities.Prefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "PrefectureRepo.FindByPrefectureID")
	defer span.End()

	prefecture := &entities.Prefecture{}
	fields, _ := prefecture.FieldMap()

	query := fmt.Sprintf("SELECT %s FROM %s WHERE prefecture_id = $1 AND deleted_at IS NULL", strings.Join(fields, ","), prefecture.TableName())

	err := database.Select(ctx, db, query, prefectureID).ScanOne(prefecture)

	switch err {
	case nil:
		return prefecture, nil
	case pgx.ErrNoRows:
		return nil, err
	default:
		return nil, fmt.Errorf("err FindByID PrefectureRepo: %w", err)
	}
}
