package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
)

type StudentRepo struct{}

func (r *StudentRepo) GetByIDForUpdate(ctx context.Context, db database.QueryExecer, studentID string) (entities.Student, error) {
	student := &entities.Student{}
	studentFieldNames, studentFieldValues := student.FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = $1 AND deleted_at is NULL
		FOR NO KEY UPDATE`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentFieldNames, ","),
		student.TableName(),
	)
	row := db.QueryRow(ctx, stmt, studentID)
	err := row.Scan(studentFieldValues...)
	if err != nil {
		return entities.Student{}, fmt.Errorf("row.Scan: %w", err)
	}
	return *student, nil
}

func (r *StudentRepo) GetByIDs(ctx context.Context, db database.QueryExecer, entitiesIDs []string) ([]*entities.Student, error) {
	var students []*entities.Student
	studentFieldNames, _ := (&entities.Student{}).FieldMap()
	stmt :=
		`
		SELECT %s
		FROM 
			%s
		WHERE 
			student_id = ANY($1) 
		`
	stmt = fmt.Sprintf(
		stmt,
		strings.Join(studentFieldNames, ","),
		(&entities.Student{}).TableName(),
	)
	rows, err := db.Query(ctx, stmt, entitiesIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		student := new(entities.Student)
		_, fieldValues := student.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		students = append(students, student)
	}
	return students, nil
}
