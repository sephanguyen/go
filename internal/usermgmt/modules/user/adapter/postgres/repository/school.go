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

type SchoolRepo struct{}

func (s *SchoolRepo) Find(ctx context.Context, db database.QueryExecer, schoolID pgtype.Int4) (*entity.School, error) {
	ctx, span := interceptors.StartSpan(ctx, "SchoolRepo.Find")
	defer span.End()

	school := &entity.School{}
	fields := database.GetFieldNames(school)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE school_id = $1", strings.Join(fields, ","), school.TableName())
	row := db.QueryRow(ctx, query, &schoolID)
	if err := row.Scan(database.GetScanFields(school, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return school, nil
}
