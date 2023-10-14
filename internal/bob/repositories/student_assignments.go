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

type StudentAssignmentRepo struct{}

func (r *StudentAssignmentRepo) FindStudentAssignmentByAssignmentID(ctx context.Context, db database.QueryExecer, ids pgtype.TextArray) ([]*entities.StudentAssignment, error) {
	ctx, span := interceptors.StartSpan(ctx, "AssignmentRepo.FindAssignmentById")
	defer span.End()
	t := &entities.StudentAssignment{}
	fields := database.GetFieldNames(t)
	query := fmt.Sprintf("SELECT %s FROM %s WHERE assignment_id = ANY ($1) ", strings.Join(fields, ","), t.TableName())
	rows, err := db.Query(ctx, query, &ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pp []*entities.StudentAssignment
	for rows.Next() {
		p := new(entities.StudentAssignment)
		if err := rows.Scan(database.GetScanFields(p, fields)...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		pp = append(pp, p)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Err")
	}
	return pp, nil
}

// UpdateStudentAssignmentStatus update student assignment status. Make sure position of studentIDs and assignment IDs doesn't affect result.
func (r *StudentAssignmentRepo) UpdateStudentAssignmentStatus(ctx context.Context, db database.QueryExecer, studentIDs pgtype.TextArray, assignmentIDs pgtype.TextArray, status pgtype.Text) error {
	ctx, span := interceptors.StartSpan(ctx, "StudentAssignmentRepo.UpdateStudentAssignmentStatus")
	defer span.End()
	statusQuery := " assignment_status = $1 "
	if status.String == entities.StudentAssignmentStatusCompleted {
		statusQuery = " assignment_status = $1, completed_at = NOW() "
	}
	query := fmt.Sprintf(`UPDATE student_assignments SET %s , updated_at = NOW()
	WHERE student_id=ANY($2) AND assignment_id = ANY($3)`, statusQuery)
	_, err := db.Exec(ctx, query, &status, &studentIDs, &assignmentIDs)
	if err != nil {
		return err
	}
	return nil
}
