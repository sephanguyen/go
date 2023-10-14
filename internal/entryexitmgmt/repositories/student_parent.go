package repositories

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	user_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/pkg/errors"
)

// StudentParentRepo provides method to work with student_parent entity
type StudentParentRepo struct{}

func (r *StudentParentRepo) GetParentIDsByStudentID(ctx context.Context, db database.QueryExecer, studentID string) ([]string, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentParentRepo.GetParentIDsByStudentID")
	defer span.End()

	studentParent := &user_entities.StudentParent{}
	query := fmt.Sprintf(`SELECT parent_id FROM %s WHERE student_id = $1 And deleted_at IS NULL`, studentParent.TableName())
	rows, err := db.Query(ctx, query, &studentID)

	if err != nil {
		return nil, fmt.Errorf("err GetParentIDsByStudentID StudentParentRepo: %w", err)
	}
	defer rows.Close()

	var parentIDs []string
	for rows.Next() {
		p := new(user_entities.StudentParent)
		if err := rows.Scan(database.GetScanFields(p, []string{"parent_id"})...); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		parentIDs = append(parentIDs, p.ParentID.String)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("err GetParentIDsByStudentID StudentParentRepo: %w", err)
	}

	return parentIDs, nil
}
