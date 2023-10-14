package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type GradeRepo struct{}

func (r *GradeRepo) GetByID(ctx context.Context, db database.QueryExecer, gradeID string) (entities.Grade, error) {
	grade := &entities.Grade{}
	gradeFieldNames, gradeFieldValues := grade.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			grade_id = $1
		FOR NO KEY UPDATE
		`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(gradeFieldNames, ","),
		grade.TableName(),
	)
	row := db.QueryRow(ctx, stmt, gradeID)
	err := row.Scan(gradeFieldValues...)
	if err != nil {
		return entities.Grade{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *grade, nil
}

func (r *GradeRepo) GetGradeNamesByGradeIDs(ctx context.Context, db database.QueryExecer, gradeIDs []string) (gradeNames []string, err error) {
	grade := &entities.Grade{}
	gradeFieldNames, gradeFieldValues := grade.FieldMap()
	stmt := `
		SELECT %s
		FROM %s
		WHERE
		    grade_id = ANY($1)`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(gradeFieldNames, ","),
		grade.TableName(),
	)
	rows, err := db.Query(ctx, stmt, gradeIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(gradeFieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		gradeNames = append(gradeNames, grade.Name.String)
	}
	return
}
