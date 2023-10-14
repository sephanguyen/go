package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

type StudentParentRepo struct {
}

func (r *StudentParentRepo) GetSiblingIDsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) ([]string, error) {
	_, span := interceptors.StartSpan(ctx, "StudentPaymentDetailRepo.GetStudentSiblingsByStudentID")
	defer span.End()

	query := `
			SELECT 
				student_id
			FROM student_parents sp
			INNER JOIN (
				SELECT parent_id from student_parents
				WHERE student_id = $1 AND deleted_at IS null
			) AS sp2
			ON sp.parent_id = sp2.parent_id where sp.student_id != $1
	`

	var siblingIDs []string
	rows, err := db.Query(ctx, query, studentID)
	if err != nil {
		return siblingIDs, err
	}

	defer rows.Close()

	for rows.Next() {
		var siblingID string

		err := rows.Scan(
			&siblingID,
		)
		if err != nil {
			return siblingIDs, fmt.Errorf("row.Scan: %w", err)
		}

		siblingIDs = append(siblingIDs, siblingID)
	}

	return siblingIDs, nil
}
