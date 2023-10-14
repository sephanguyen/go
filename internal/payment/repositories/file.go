package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

type FileRepo struct {
}

func (r *FileRepo) GetByFileName(ctx context.Context, db database.QueryExecer, fileName string) (entities.File, error) {
	file := &entities.File{}
	fileFieldNames, fileFieldValues := file.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			file_name = $1
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(fileFieldNames, ","),
		file.TableName(),
	)
	row := db.QueryRow(ctx, stmt, fileName)
	err := row.Scan(fileFieldValues...)
	if err != nil {
		return entities.File{}, err
	}
	return *file, nil
}

// Create creates File entity
func (r *FileRepo) Create(ctx context.Context, db database.QueryExecer, e *entities.File) error {
	ctx, span := interceptors.StartSpan(ctx, "File.Create")
	defer span.End()

	now := time.Now()
	if err := multierr.Combine(
		e.UpdatedAt.Set(now),
		e.CreatedAt.Set(now),
	); err != nil {
		return fmt.Errorf("multierr.Combine UpdatedAt.Set CreatedAt.Set: %w", err)
	}

	cmdTag, err := database.InsertExcept(ctx, e, []string{"resource_path"}, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert File: %w", err)
	}

	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert FileRepo: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}
