package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type StudentEnrollmentStatusHistoryRepo struct{}

func (u *StudentEnrollmentStatusHistoryRepo) GetStatusHistoryByStudentIDsAndLocationID(ctx context.Context, db database.QueryExecer, studentIDs []string, locationID string) (domain.StudentEnrollmentStatusHistories, error) {
	ctx, span := interceptors.StartSpan(ctx, "StudentEnrollmentStatusHistoryRepo.GetStatusHistoryByStudentIDsAndLocationID")
	defer span.End()

	studentEnrollmentHistory := &StudentEnrollmentStatusHistory{}
	fields, values := studentEnrollmentHistory.FieldMap()

	query := fmt.Sprintf(`SELECT %s FROM %s 
				WHERE student_id = ANY($1) 
				AND location_id = $2 
				AND enrollment_status = 'STUDENT_ENROLLMENT_STATUS_ENROLLED' 
				AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		studentEnrollmentHistory.TableName())

	rows, err := db.Query(ctx, query, studentIDs, locationID)
	if err != nil {
		return nil, fmt.Errorf("db.Query: %w", err)
	}
	defer rows.Close()

	var result domain.StudentEnrollmentStatusHistories
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		result = append(result, studentEnrollmentHistory.ToStudentEnrollmentStatusHistoryDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return result, nil
}
