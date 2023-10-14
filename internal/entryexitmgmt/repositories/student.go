package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type StudentRepo struct{}

func (r *StudentRepo) FindByID(ctx context.Context, db database.QueryExecer, studentID string) (*entities.Student, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentRepo.FindByID")
	defer span.End()

	e := &entities.Student{}
	fields := database.GetFieldNames(e)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE student_id = $1", strings.Join(fields, ","), e.TableName())
	row := db.QueryRow(ctx, query, &studentID)
	if err := row.Scan(database.GetScanFields(e, fields)...); err != nil {
		return nil, fmt.Errorf("row.Scan: %w", err)
	}

	return e, nil
}
