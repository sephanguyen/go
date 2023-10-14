package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
)

type PrefectureRepo struct{}

func (r *PrefectureRepo) GetByPrefectureID(ctx context.Context, db database.QueryExecer, prefectureID pgtype.Text) (*entity.Prefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.GetByPrefectureCode")
	defer span.End()

	prefectureEnt := &entity.Prefecture{}
	fields := database.GetFieldNames(prefectureEnt)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE prefecture_id = $1 and deleted_at IS NULL", strings.Join(fields, ","), prefectureEnt.TableName())
	row := db.QueryRow(ctx, query, &prefectureID)
	if err := row.Scan(database.GetScanFields(prefectureEnt, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return prefectureEnt, nil
}

func (r *PrefectureRepo) GetByPrefectureCode(ctx context.Context, db database.QueryExecer, prefectureCode pgtype.Text) (*entity.Prefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "ParentRepo.GetByPrefectureCode")
	defer span.End()

	prefectureEnt := &entity.Prefecture{}
	fields := database.GetFieldNames(prefectureEnt)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE prefecture_code = $1 and deleted_at IS NULL", strings.Join(fields, ","), prefectureEnt.TableName())
	row := db.QueryRow(ctx, query, &prefectureCode)
	if err := row.Scan(database.GetScanFields(prefectureEnt, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return prefectureEnt, nil
}

func (r *PrefectureRepo) GetByPrefectureIDs(ctx context.Context, db database.QueryExecer, prefectureIDs pgtype.TextArray) ([]*entity.Prefecture, error) {
	ctx, span := interceptors.StartSpan(ctx, "PrefectureRepo.GetByPrefectureCodes")
	defer span.End()

	prefecture := &entity.Prefecture{}
	fields := database.GetFieldNames(prefecture)
	stmt := fmt.Sprintf("SELECT %s FROM %s WHERE prefecture_id = ANY($1) AND deleted_at IS NULL", strings.Join(fields, ","), prefecture.TableName())

	rows, err := db.Query(ctx, stmt, &prefectureIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefectures := make([]*entity.Prefecture, 0)
	for rows.Next() {
		prefecture := &entity.Prefecture{}
		if err := rows.Scan(database.GetScanFields(prefecture, fields)...); err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		prefectures = append(prefectures, prefecture)
	}

	return prefectures, nil
}
