package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type StudentParentRepo struct{}

func (s *StudentParentRepo) GetStudentParents(ctx context.Context, db database.QueryExecer, studentIDs []string) ([]domain.StudentParent, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.GetStudentParents")
	defer span.End()

	dto := &StudentParent{}
	fields, values := dto.FieldMap()

	query := fmt.Sprintf(`
		SELECT %s FROM %s 
		WHERE student_id = ANY($1) 
		AND deleted_at IS NULL`,
		strings.Join(fields, ", "),
		dto.TableName(),
	)

	rows, err := db.Query(ctx, query, studentIDs)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	studentParents := make([]domain.StudentParent, 0)
	for rows.Next() {
		err := rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		studentParents = append(studentParents, dto.ToStudentParentDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return studentParents, nil
}
